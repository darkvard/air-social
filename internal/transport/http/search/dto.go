package search

import (
	"air-social/internal/domain/common"
	"air-social/internal/domain/search"
	"air-social/internal/transport/http/shared"
)

type CursorQueryParams struct {
	Query  string `form:"query" binding:"required,min=1,max=100"`
	Cursor int64  `form:"cursor" binding:"omitempty,min=0"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
	Sort   string `form:"sort,default=latest" binding:"omitempty,oneof=latest oldest"`
}

func (q CursorQueryParams) toUserParams() search.UsersParams {
	return search.UsersParams{
		Search: q.Query,
		Query:  common.CursorQueryParams[int64]{Cursor: q.Cursor, Limit: q.Limit, Sort: q.Sort},
	}
}

func (q CursorQueryParams) toPostParams() {
	// todo
}

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar"`
	Verified bool   `json:"verified"`
}

type MetaCursor = shared.MetaCursor

type CursorPaginatedResponse[T any] = shared.CursorPaginatedResponse[T]
