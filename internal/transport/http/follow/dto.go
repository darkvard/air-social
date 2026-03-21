package follow

import (
	"air-social/internal/domain/common"
)

type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type QueryParams struct {
	Page  int    `form:"page,default=1" binding:"min=1"`
	Limit int    `form:"limit,default=10" binding:"min=1,max=100"`
	Sort  string `form:"sort" binding:"omitempty,oneof=latest oldest name_asc name_desc"`
}

func (q QueryParams) ToDomain() common.OffsetQueryParams {
	return common.OffsetQueryParams{
		Page:  q.Page,
		Limit: q.Limit,
		Sort:  q.Sort,
	}
}

type UserFollowResponse struct {
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	Avatar      string `json:"avatar"`
	IsVerified  bool   `json:"is_verified"`
	IsFollowing bool   `json:"is_following"`
	IsFollower  bool   `json:"is_follower"`
}

type FollowListResponse struct {
	Data       []UserFollowResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

func toFollowListResponse(result common.OffsetPaginatedResult[UserFollowResponse]) FollowListResponse {
	return FollowListResponse{
		Data:       result.Data,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}
}
