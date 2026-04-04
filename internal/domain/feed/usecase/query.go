package usecase

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/internal/domain/feed"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type PostFetcher interface {
	GetPostsByIDs(ctx context.Context, postIDs []int64, viewerID int64) ([]*post.Post, error)
}

type QueryDeps struct {
	CacheProvider feed.Cache
	PostFetcher   PostFetcher
}

type queryUseCase struct {
	deps QueryDeps
}

func NewQueryUseCase(deps QueryDeps) *queryUseCase {
	return &queryUseCase{deps: deps}
}

func (u *queryUseCase) GetNewsfeed(ctx context.Context, viewerID int64, params common.CursorQueryParams[int64]) (common.CursorPaginatedResult[*post.Post, int64], error) {
	params.NormalizePagination()
	var empty common.CursorPaginatedResult[*post.Post, int64]

	// Fetch limit+1 from Cache to detect whether a next page exists.
	postIDs, err := u.deps.CacheProvider.GetFeedPostIDs(ctx, viewerID, params.Cursor, params.GetFetchLimit())
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	if len(postIDs) == 0 {
		return empty, nil
	}

	// Trim to limit before DB fetch — the extra element was only needed to detect hasNextPage.
	hasNextPage := len(postIDs) > params.Limit
	if hasNextPage {
		postIDs = postIDs[:params.Limit]
	}

	// Fetch from DB: Get post details.
	// WARNING: SQL IN clause does not preserve order; re-ordering is done below.
	posts, err := u.deps.PostFetcher.GetPostsByIDs(ctx, postIDs, viewerID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	// Hydration & Re-ordering (O(N)):
	// Build a map for O(1) lookups, then walk postIDs in Redis order.
	postsMap := make(map[int64]*post.Post, len(posts))
	for _, p := range posts {
		postsMap[p.ID] = p
	}

	orderedPosts := make([]*post.Post, 0, len(postIDs))
	for _, id := range postIDs {
		if p, exists := postsMap[id]; exists {
			orderedPosts = append(orderedPosts, p)
		}
	}

	// nextCursor = CreatedAt.UnixMilli() of the last post — must match the Redis score.
	// Cannot use NewCursorPaginatedResult here: Post.GetCursor() returns ID, not timestamp.
	var nextCursor int64
	if hasNextPage && len(orderedPosts) > 0 {
		nextCursor = orderedPosts[len(orderedPosts)-1].CreatedAt.UnixMilli()
	}

	return common.CursorPaginatedResult[*post.Post, int64]{
		Data:        orderedPosts,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
