package stats

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"air-social/pkg"
)

// UseCase implements "Write-Behind Caching" & "Base + Delta" architecture to prevent DB Write Storms.
//
// Logic:
// 1. Write: User actions (Like/Comment/Share) push events to RabbitMQ instead of DB.
// 2. Cache: Workers process events and update delta counters (+/-) in Redis.
// 3. Read: Fetch DB (Base) + Redis (Delta) concurrently -> Final count = max(0, Base + Delta).
// 4. Sync: Cronjob runs every ...s to Bulk Upsert Redis deltas to PostgreSQL, then subtracts synced values via Lua.
type UseCase interface {
	SyncPostStats(ctx context.Context) error
	SyncCommentStats(ctx context.Context) error

	GetPostsStats(ctx context.Context, postIDs []int64) (map[int64]PostStats, error)
	GetCommentsStats(ctx context.Context, commentIDs []int64) (map[int64]CommentStats, error)

	ReconcilePostStats(ctx context.Context, postIDs []int64) error
	ReconcileCommentStats(ctx context.Context, commentIDs []int64) error
}

type Deps struct {
	Repo  Repository
	Cache Cache
}

type usecase struct {
	repo  Repository
	cache Cache
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{repo: deps.Repo, cache: deps.Cache}
}

// SyncPostStats executes the Write-Behind sync flow for posts.
// Concept: Concurrent Fetch -> Deduplicate IDs -> Assemble Batch -> Bulk Upsert
func (u *usecase) SyncPostStats(ctx context.Context) error {
	states := []string{StatePostLikes, StatePostComments, StatePostShares}
	maps, err := u.fetchMultipleHashes(ctx, states...)
	if err != nil {
		return pkg.NewError(err, "failed to fetch post stats from cache")
	}

	uniqueIDs := u.extractUniqueIDs(maps...)
	if len(uniqueIDs) == 0 {
		return nil
	}

	size := len(uniqueIDs)
	params := PostParams{
		IDs:      uniqueIDs,
		Likes:    make([]int64, size),
		Comments: make([]int64, size),
		Shares:   make([]int64, size),
	}

	for i, id := range uniqueIDs {
		params.Likes[i] = maps[0][id]
		params.Comments[i] = maps[1][id]
		params.Shares[i] = maps[2][id]
	}

	if err := u.repo.BulkUpsertPostStats(ctx, params); err != nil {
		go func(ids []int64) {
			_ = u.ReconcilePostStats(context.Background(), ids)
		}(uniqueIDs)
		return pkg.OrInternalError(err)
	}

	u.clearMultipleCaches(states, maps)
	return nil
}

// SyncCommentStats executes the Write-Behind sync flow for comments.
func (u *usecase) SyncCommentStats(ctx context.Context) error {
	states := []string{StateCommentLikes, StateCommentReplies}
	maps, err := u.fetchMultipleHashes(ctx, states...)
	if err != nil {
		return pkg.NewError(err, "failed to fetch comment stats from cache")
	}

	uniqueIDs := u.extractUniqueIDs(maps...)
	if len(uniqueIDs) == 0 {
		return nil
	}

	size := len(uniqueIDs)
	params := CommentParams{
		IDs:     uniqueIDs,
		Likes:   make([]int64, size),
		Replies: make([]int64, size),
	}

	for i, id := range uniqueIDs {
		params.Likes[i] = maps[0][id]
		params.Replies[i] = maps[1][id]
	}

	if err := u.repo.BulkUpsertCommentStats(ctx, params); err != nil {
		go func(ids []int64) {
			_ = u.ReconcileCommentStats(context.Background(), ids)
		}(uniqueIDs)
		return pkg.OrInternalError(err)
	}

	u.clearMultipleCaches(states, maps)
	return nil
}

// ReconcilePostStats acts as an automatic fallback when the primary sync flow fails (e.g., Redis down, bulk upsert error).
// It directly executes SQL queries to recalculate post stats from the source of truth tables and overwrites the DB to ensure data integrity.
func (u *usecase) ReconcilePostStats(ctx context.Context, postIDs []int64) error {
	if len(postIDs) == 0 {
		return nil
	}
	if err := u.repo.ReconcilePostStats(ctx, postIDs); err != nil {
		return pkg.NewError(err, "failed to reconcile post stats")
	}
	return nil
}

// ReconcileCommentStats acts as an automatic fallback when the primary sync flow fails (e.g., Redis down, bulk upsert error).
// It directly executes SQL queries to recalculate comment stats from the source of truth tables and overwrites the DB to ensure data integrity.
func (u *usecase) ReconcileCommentStats(ctx context.Context, commentIDs []int64) error {
	if len(commentIDs) == 0 {
		return nil
	}
	if err := u.repo.ReconcileCommentStats(ctx, commentIDs); err != nil {
		return pkg.NewError(err, "failed to reconcile comment stats")
	}
	return nil
}

// GetPostsStats fetches base stats from PostgreSQL and merges them with live offsets from Redis.
//
// Architecture / Concept:
// - Database (Base): Contains persistent data but may be delayed (up to time.s) due to the background syncer.
// - Redis (Delta): Contains the most recent, unsynced interactions (+/- offsets).
//
// Solution:
// - Fetch both Base and Delta concurrently using errgroup to minimize latency.
// - Merge Data: Final Count = GREATEST(0, Base_DB + Offset_Redis).
func (u *usecase) GetPostsStats(ctx context.Context, postIDs []int64) (map[int64]PostStats, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	var dbStats []PostStats
	var cacheOffsets []map[int64]int64

	g, gCtx := errgroup.WithContext(ctx)

	// get from db
	g.Go(func() error {
		var err error
		dbStats, err = u.repo.GetPostsStats(gCtx, postIDs)
		return err
	})

	// get from cache
	g.Go(func() error {
		var err error
		states := []string{StatePostLikes, StatePostComments, StatePostShares}
		cacheOffsets, err = u.fetchMultipleOffsets(gCtx, states, postIDs)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, pkg.NewError(err, "failed to fetch post stats")
	}

	// merge
	result := make(map[int64]PostStats, len(postIDs))
	for _, st := range dbStats {
		result[st.PostID] = st
	}

	for _, id := range postIDs {
		st, exists := result[id]
		if !exists {
			st = PostStats{PostID: id}
		}

		// [0]=Likes, [1]=Comments, [2]=Shares
		likes := max(st.LikesCount+int32(cacheOffsets[0][id]), 0)
		comments := max(st.CommentsCount+int32(cacheOffsets[1][id]), 0)
		shares := max(st.SharesCount+int32(cacheOffsets[2][id]), 0)

		st.LikesCount = likes
		st.CommentsCount = comments
		st.SharesCount = shares

		result[id] = st
	}

	return result, nil
}

// GetCommentsStats fetches base stats from PostgreSQL and merges them with live offsets from Redis.
// It follows the exact same Base + Delta aggregation strategy as Posts.
func (u *usecase) GetCommentsStats(ctx context.Context, commentIDs []int64) (map[int64]CommentStats, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}

	var dbStats []CommentStats
	var cacheOffsets []map[int64]int64

	g, gCtx := errgroup.WithContext(ctx)

	// get from db
	g.Go(func() error {
		var err error
		dbStats, err = u.repo.GetCommentsStats(gCtx, commentIDs)
		return err
	})

	// get from cache
	g.Go(func() error {
		var err error
		states := []string{StateCommentLikes, StateCommentReplies}
		cacheOffsets, err = u.fetchMultipleOffsets(gCtx, states, commentIDs)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, pkg.NewError(err, "failed to fetch comment stats")
	}

	// merge
	result := make(map[int64]CommentStats, len(commentIDs))
	for _, st := range dbStats {
		result[st.CommentID] = st
	}

	for _, id := range commentIDs {
		st, exists := result[id]
		if !exists {
			st = CommentStats{CommentID: id}
		}

		// [0]=Likes, [1]=Replies
		likes := max(st.LikesCount+int32(cacheOffsets[0][id]), 0)
		replies := max(st.RepliesCount+int32(cacheOffsets[1][id]), 0)

		st.LikesCount = likes
		st.RepliesCount = replies

		result[id] = st
	}

	return result, nil
}

func (u *usecase) fetchMultipleHashes(ctx context.Context, states ...string) ([]map[int64]int64, error) {
	results := make([]map[int64]int64, len(states))
	g, gCtx := errgroup.WithContext(ctx)

	for i, state := range states {
		g.Go(func() error {
			data, err := u.cache.GetStatsHash(gCtx, state)
			if err != nil {
				return err
			}
			results[i] = data
			return nil
		})
	}

	return results, g.Wait()
}

func (u *usecase) extractUniqueIDs(maps ...map[int64]int64) []int64 {
	uniqueSet := make(map[int64]struct{})
	for _, m := range maps {
		for id := range m {
			uniqueSet[id] = struct{}{}
		}
	}

	if len(uniqueSet) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(uniqueSet))
	for id := range uniqueSet {
		ids = append(ids, id)
	}
	return ids
}

func (u *usecase) clearMultipleCaches(states []string, maps []map[int64]int64) {
	var wg sync.WaitGroup
	wg.Add(len(states))

	for i := range states {
		go func() {
			defer wg.Done()
			_ = u.cache.ClearSyncedFields(context.Background(), states[i], maps[i])
		}()
	}
	wg.Wait()
}

func (u *usecase) fetchMultipleOffsets(ctx context.Context, states []string, ids []int64) ([]map[int64]int64, error) {
	results := make([]map[int64]int64, len(states))
	g, gCtx := errgroup.WithContext(ctx)

	for i, state := range states {
		g.Go(func() error {
			offsets, err := u.cache.GetStatsOffsets(gCtx, state, ids)
			if err != nil {
				return err
			}

			if offsets == nil {
				offsets = make(map[int64]int64)
			}

			results[i] = offsets
			return nil
		})
	}

	return results, g.Wait()
}
