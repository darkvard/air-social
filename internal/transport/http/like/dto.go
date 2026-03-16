package like

import (
	"air-social/internal/domain/common"
	"air-social/internal/domain/like"
)

type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type MetaCursor struct {
	NextCursor  int64 `json:"next_cursor"`
	HasNextPage bool  `json:"has_next_page"`
}

type CursorPaginatedResponse[T any] struct {
	Data []T        `json:"data"`
	Meta MetaCursor `json:"meta"`
}

type LikerResponse struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
}

type CursorQueryParams struct {
	Cursor int64  `form:"cursor" binding:"omitempty,min=0"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
	Sort   string `form:"sort,default=latest" binding:"omitempty,oneof=latest oldest"`
}

func (q CursorQueryParams) ToDomain(targetID int64) like.GetCursorParams {
	return like.GetCursorParams{
		TargetID: targetID,
		Query: common.CursorQueryParams[int64]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
			Sort:   q.Sort,
		},
	}
}
