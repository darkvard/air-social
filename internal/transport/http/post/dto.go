package post

import (
	"time"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	"air-social/internal/transport/http/shared"
)

// PathIDParam is re-exported from shared for convenience within this package.
type PathIDParam = shared.PathIDParam

type CreateRequest struct {
	OriginalPostID *int64           `json:"original_post_id" binding:"omitempty,gt=0"`
	Content        string           `json:"content" binding:"required_without=Media"`
	Visibility     string           `json:"visibility" binding:"omitempty,oneof=public followers private"`
	Media          []MediaItemInput `json:"media" binding:"omitempty,max=10,dive"`
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

func (q CursorQueryParams) ToGetSharersParams(postID int64) post.GetSharersParams {
	return post.GetSharersParams{
		PostID: postID,
		Query: common.CursorQueryParams[int64]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
			Sort:   q.Sort,
		},
	}
}

type CreateResponse struct {
	ID             int64               `json:"id"`
	OriginalPostID *int64              `json:"original_post_id,omitempty"`
	Content        string              `json:"content"`
	Visibility     string              `json:"visibility"`
	CreatedAt      time.Time           `json:"created_at"`
	Media          []MediaItemResponse `json:"media"`
}

type PostResponse struct {
	ID             int64                `json:"id"`
	OriginalPostID *int64               `json:"original_post_id,omitempty"`
	Content        string               `json:"content"`
	Visibility     string               `json:"visibility"`
	LikesCount     int32                `json:"likes_count"`
	CommentsCount  int32                `json:"comments_count"`
	SharesCount    int32                `json:"shares_count"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Media          []MediaItemResponse  `json:"media"`
	IsLiked        *bool                `json:"is_liked,omitempty"`
	User           *shared.UserResponse `json:"author,omitempty"`
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

type SharerResponse struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
}

type MetaCursor = shared.MetaCursor[int64]

type CursorPaginatedResponse[T any] = shared.CursorPaginatedResponse[T, int64]

type UserResponse = shared.UserResponse

// ToMediaItems maps domain media slice to MediaItemResponse slice.
func ToMediaItems(media []post.Media, resolveURL func(string) string) []MediaItemResponse {
	result := make([]MediaItemResponse, len(media))
	for i, m := range media {
		result[i] = MediaItemResponse{
			ID:        m.ID,
			URL:       resolveURL(m.MediaKey),
			MediaType: m.MediaType,
			Width:     m.Metadata.Width,
			Height:    m.Metadata.Height,
			Duration:  m.Metadata.Duration,
			FileName:  m.Metadata.FileName,
		}
	}
	return result
}

func ToPostResponse(p *post.Post, resolveURL func(string) string) PostResponse {
	if p == nil {
		return PostResponse{}
	}

	var author *UserResponse
	if p.Author != nil {
		author = &UserResponse{
			ID:         p.Author.ID,
			Fullname:   p.Author.FullName,
			Avatar:     resolveURL(p.Author.Avatar),
			IsVerified: p.Author.IsVerified,
		}
	}

	return PostResponse{
		ID:             p.ID,
		OriginalPostID: p.OriginalPostID,
		Content:        p.Content,
		Visibility:     string(p.Visibility),
		LikesCount:     p.Stat.LikesCount,
		CommentsCount:  p.Stat.CommentsCount,
		SharesCount:    p.Stat.SharesCount,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		Media:          ToMediaItems(p.Media, resolveURL),
		User:           author,
		IsLiked:        p.IsLiked,
	}
}
