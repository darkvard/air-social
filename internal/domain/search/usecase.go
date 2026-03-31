package search

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type UseCase interface {
	SearchUsers(ctx context.Context, params UsersParams) (common.CursorPaginatedResult[User, int64], error)

	SearchPosts(ctx context.Context, params PostsParams) (common.CursorPaginatedResult[Post, int64], error)
}

type usecase struct {
	repo Repository
}

func NewUseCase(repo Repository) *usecase {
	return &usecase{repo: repo}
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

func (u *usecase) SearchPosts(ctx context.Context, params PostsParams) (common.CursorPaginatedResult[Post, int64], error) {
	return common.CursorPaginatedResult[Post, int64]{}, nil
}
