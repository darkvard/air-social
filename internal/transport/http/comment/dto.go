package comment

import "time"

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

type CommentResponse struct {
	ID        int64               `json:"id"`
	Content   string              `json:"content"`
	ParentID  *int64              `json:"parent_id,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	Media     []MediaItemResponse `json:"media,omitempty"`
}

type UpdateRequest struct {
	Content *string          `json:"content" binding:"omitempty,max=1000"`
	Media   []MediaItemInput `json:"media" binding:"omitempty,max=4,dive"`
}
