package search

import "context"

type Repository interface {
	SearchUsers(ctx context.Context, params UsersParams) ([]User, error)

	SearchPosts(ctx context.Context, params PostsParams) ([]Post, error)
}
