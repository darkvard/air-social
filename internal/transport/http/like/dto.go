package like

import (
	"air-social/internal/domain/common"
	"air-social/internal/domain/like"
	"air-social/internal/transport/http/shared"
)

type PathIDParam = shared.PathIDParam
type MetaCursor = shared.MetaCursor
type CursorPaginatedResponse[T any] = shared.CursorPaginatedResponse[T]

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
