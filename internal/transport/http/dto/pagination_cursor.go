package dto

import "air-social/internal/domain"

type CursorPaginationQuery struct {
	Cursor int64 `form:"cursor" binding:"omitempty,min=0"`
	Limit  int   `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
}

func (q CursorPaginationQuery) ToDomain() domain.CursorQueryParams {
	return domain.CursorQueryParams{
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}
}

type CursorPaginatedResponse[T any] struct {
	Data []T        `json:"data"`
	Meta MetaCursor `json:"meta"`
}

type MetaCursor struct {
	NextCursor  int64 `json:"next_cursor"`
	HasNextPage bool  `json:"has_next_page"`
}

func NewCursorPaginatedResponse[T any](result domain.CursorPaginatedResult[T]) CursorPaginatedResponse[T] {
	return CursorPaginatedResponse[T]{
		Data: result.Data,
		Meta: MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}
