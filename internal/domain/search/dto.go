package search

import "air-social/internal/domain/common"

type UsersParams struct {
	Search string
	Query  common.CursorQueryParams[int64]
}

type PostsParams struct {
	Search string
	Query  common.CursorQueryParams[int64]
}
