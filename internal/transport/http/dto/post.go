package dto

type CreatePostRequest struct {
	Content    string           `json:"content" binding:"required_without=Media"`
	Visibility string           `json:"visibility" binding:"omitempty,oneof=public followers private"`
	Media      []MediaItemInput `json:"media" binding:"omitempty,max=10,dive"`
}

type MediaItemInput struct {
	MediaKey  string `json:"media_key" binding:"required"`
	MediaType string `json:"media_type" binding:"required,oneof=image video audio document"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  int    `json:"duration"`
	Size      int64  `json:"size"`
	FileName  string `json:"file_name"`
}

type UpdatePostRequest struct {
	Content    *string `json:"content" binding:"omitempty"`
	Visibility *string `json:"visibility" binding:"omitempty,oneof=public followers private"`
}

type PostResponse struct {
	ID            int64               `json:"id"`
	Content       string              `json:"content"`
	Visibility    string              `json:"visibility"`
	Version       int                 `json:"version"`
	LikesCount    int                 `json:"likes_count"`
	CommentsCount int                 `json:"comments_count"`
	SharesCount   int                 `json:"shares_count"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
	IsLiked       bool                `json:"is_liked"`
	User          UserCompactResponse `json:"user"`
	Media         []MediaItemResponse `json:"media"`
}

type MediaItemResponse struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	FileName  string `json:"file_name,omitempty"`
}

type UserCompactResponse struct {
	ID       int64  `json:"id"`
	Fullname string `json:"full_name"`
	Avatar   string `json:"avatar"`
}
