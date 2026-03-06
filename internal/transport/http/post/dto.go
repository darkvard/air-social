package post

import (
	"time"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
)

type PathIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type CreateRequest struct {
	Content    string           `json:"content" binding:"required_without=Media"`
	Visibility string           `json:"visibility" binding:"omitempty,oneof=public followers private"`
	Media      []MediaItemInput `json:"media" binding:"omitempty,max=10,dive"`
}

type MediaItemInput struct {
	MediaKey  string `json:"media_key" binding:"required"`
	MediaType string `json:"media_type" binding:"required,oneof=image video audio document"`
	Width     int32  `json:"width" binding:"omitempty,min=0"`
	Height    int32  `json:"height" binding:"omitempty,min=0"`
	Duration  int32  `json:"duration" binding:"omitempty,min=0"`
	Size      int64  `json:"size" binding:"required,gt=0"`
	FileName  string `json:"file_name" binding:"required,max=255"`
}

type UpdateRequest struct {
	Content    *string          `json:"content" binding:"omitempty"`
	Visibility *string          `json:"visibility" binding:"omitempty,oneof=public followers private"`
	Media      []MediaItemInput `json:"media" binding:"omitempty,max=10,dive"`
}

type CursorQueryParams struct {
	Cursor int64  `form:"cursor" binding:"omitempty,min=0"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
	Sort   string `form:"sort,default=latest" binding:"omitempty,oneof=latest oldest"`
}

func (q CursorQueryParams) ToDomain(userID int64) post.GetCursorParams {
	return post.GetCursorParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
			Sort:   q.Sort,
		},
	}
}

type CreateResponse struct {
	ID         int64               `json:"id"`
	Content    string              `json:"content"`
	Visibility string              `json:"visibility"`
	CreatedAt  time.Time           `json:"created_at"`
	Media      []MediaItemResponse `json:"media"`
}

type PostResponse struct {
	ID            int64               `json:"id"`
	Content       string              `json:"content"`
	Visibility    string              `json:"visibility"`
	LikesCount    int32               `json:"likes_count"`
	CommentsCount int32               `json:"comments_count"`
	SharesCount   int32               `json:"shares_count"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	Media         []MediaItemResponse `json:"media"`
	IsLiked       *bool               `json:"is_liked,omitempty"`
	User          *UserResponse       `json:"author,omitempty"`
}

type MediaItemResponse struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Width     int32  `json:"width,omitempty"`
	Height    int32  `json:"height,omitempty"`
	Duration  int32  `json:"duration,omitempty"`
	FileName  string `json:"file_name,omitempty"`
}

type UserResponse struct {
	ID         int64  `json:"id"`
	Fullname   string `json:"full_name"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
}

type CursorPaginatedResponse[T any] struct {
	Data []T        `json:"data"`
	Meta MetaCursor `json:"meta"`
}

type MetaCursor struct {
	NextCursor  int64 `json:"next_cursor"`
	HasNextPage bool  `json:"has_next_page"`
}
