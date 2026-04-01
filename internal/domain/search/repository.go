package search

import "context"

type Repository interface {
	SearchUsers(ctx context.Context, params UsersParams) ([]User, error)

	SearchPostIDs(ctx context.Context, params PostsParams) ([]int64, error)
}
