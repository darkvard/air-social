package search

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type PostFetcher interface {
	GetPostsByIDs(ctx context.Context, postIDs []int64, viewerID int64) ([]*post.Post, error)
}

type UseCase interface {
	SearchUsers(ctx context.Context, params UsersParams) (common.CursorPaginatedResult[User, int64], error)

	SearchPosts(ctx context.Context, params PostsParams) (common.CursorPaginatedResult[*post.Post, int64], error)
}

type usecase struct {
	repo        Repository
	postFetcher PostFetcher
}

func NewUseCase(repo Repository, postFetcher PostFetcher) *usecase {
	return &usecase{repo: repo, postFetcher: postFetcher}
}

func (u *usecase) SearchUsers(ctx context.Context, params UsersParams) (common.CursorPaginatedResult[User, int64], error) {
	var empty common.CursorPaginatedResult[User, int64]
	params.Query.NormalizePagination()

	users, err := u.repo.SearchUsers(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return common.NewCursorPaginatedResult(users, params.Query.Limit), nil
}

func (u *usecase) SearchPosts(ctx context.Context, params PostsParams) (common.CursorPaginatedResult[*post.Post, int64], error) {
	var empty common.CursorPaginatedResult[*post.Post, int64]
	params.Query.NormalizePagination()

	// find matching post IDs via FTS
	ids, err := u.repo.SearchPostIDs(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	if len(ids) == 0 {
		return common.CursorPaginatedResult[*post.Post, int64]{Data: []*post.Post{}}, nil
	}

	// detect next page, trim IDs
	hasNextPage := len(ids) > params.Query.Limit
	if hasNextPage {
		ids = ids[:params.Query.Limit]
	}

	// hydrate full post data (stats, likes, media, author) via PostFetcher 
	posts, err := u.postFetcher.GetPostsByIDs(ctx, ids, params.ViewerID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	// re-order posts to match FTS ordering (SQL IN does not preserve order)
	postsMap := make(map[int64]*post.Post, len(posts))
	for _, p := range posts {
		postsMap[p.ID] = p
	}

	ordered := make([]*post.Post, 0, len(ids))
	for _, id := range ids {
		if p, exists := postsMap[id]; exists {
			ordered = append(ordered, p)
		}
	}

	var nextCursor int64
	if hasNextPage && len(ordered) > 0 {
		nextCursor = ordered[len(ordered)-1].ID
	}

	return common.CursorPaginatedResult[*post.Post, int64]{
		Data:        ordered,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
