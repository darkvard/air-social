package stats

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"air-social/internal/domain/stats/cache"
	"air-social/pkg"
)

type UseCase interface {
	SyncPostStats(ctx context.Context) error
	SyncCommentStats(ctx context.Context) error
}

type Deps struct {
	Repo  Repository
	Cache cache.Provider
}

type usecase struct {
	repo  Repository
	cache cache.Provider
}

func NewUseCase(deps Deps) UseCase {
	return &usecase{repo: deps.Repo, cache: deps.Cache}
}

// SyncPostStats executes the Write-Behind sync flow for posts.
// Concept: Concurrent Fetch -> Deduplicate IDs -> Assemble Batch -> Bulk Upsert
func (u *usecase) SyncPostStats(ctx context.Context) error {
	states := []string{cache.StatePostLikes, cache.StatePostComments, cache.StatePostShares}
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
		return pkg.OrInternalError(err)
	}

	u.clearMultipleCaches(states, maps)
	return nil
}

// SyncCommentStats executes the Write-Behind sync flow for comments.
func (u *usecase) SyncCommentStats(ctx context.Context) error {
	states := []string{cache.StateCommentLikes, cache.StateCommentReplies}
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
		return pkg.OrInternalError(err)
	}

	u.clearMultipleCaches(states, maps)
	return nil
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
		i := i // Capture
		go func() {
			defer wg.Done()
			_ = u.cache.ClearSyncedFields(context.Background(), states[i], maps[i])
		}()
	}
	wg.Wait()
}
