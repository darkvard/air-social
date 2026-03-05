package comment

import (
	"time"

	"air-social/internal/domain/comment"
	"air-social/internal/domain/common"
)

type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type MediaItemInput struct {
	MediaKey  string `json:"media_key" binding:"required"`
	MediaType string `json:"media_type" binding:"required,oneof=image video audio document"`
}

type CreateRequest struct {
	Content  string           `json:"content" binding:"required,max=1000"`
	ParentID *int64           `json:"parent_id" binding:"omitempty,min=1"`
	Media    []MediaItemInput `json:"media" binding:"max=4"`
}

type MediaItemResponse struct {
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
}

type UpdateRequest struct {
	Content *string          `json:"content" binding:"omitempty,max=1000"`
	Media   []MediaItemInput `json:"media" binding:"omitempty,max=4,dive"`
}

type CursorQueryParams struct {
	Cursor int64  `form:"cursor" binding:"omitempty,min=0"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
	Sort   string `form:"sort,default=latest" binding:"omitempty,oneof=latest oldest"`
}

func (q CursorQueryParams) ToDomain(userID int64) comment.GetCursorParams {
	return comment.GetCursorParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
		},
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

type CommentResponse struct {
	ID        int64               `json:"id"`
	Content   string              `json:"content"`
	ParentID  *int64              `json:"parent_id,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	Media     []MediaItemResponse `json:"media,omitempty"`
	User      *UserResponse       `json:"author,omitempty"`
	IsLiked   *bool               `json:"is_liked"`
}

type UserResponse struct {
	ID         int64  `json:"id"`
	Fullname   string `json:"full_name"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
}
