package domain

import (
	"context"
	"time"
)

const (
	VisibilityPublic    PostVisibility = "public"
	VisibilityFollowers PostVisibility = "followers"
	VisibilityPrivate   PostVisibility = "private"
)

type PostRepository interface {
	GetByID(ctx context.Context, id int64) (*Post, error)
	GetByUserID(ctx context.Context, userID int64, cursor int64, limit int) ([]Post, error)
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int64) error
	IsOwner(ctx context.Context, postID int64, userID int64) (bool, error)
}

type PostVisibility string

type Post struct {
	ID         int64
	UserID     int64
	Content    string
	Visibility PostVisibility
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time

	Media []PostMedia
	User  *UserSummary
}

type PostMedia struct {
	ID        int64
	PostID    int64
	MediaKey  string
	MediaType string
	Metadata  PostMediaMetadata
	CreatedAt time.Time
}

type PostMediaMetadata struct {
	Width    int
	Height   int
	Duration int
	Size     int64
	FileName string
}

type CreatePostParams struct {
	UserID     int64
	Content    string
	Visibility PostVisibility
	Media      []PostMediaParams
}

type PostMediaParams struct {
	MediaKey  string
	MediaType string
	Width     int
	Height    int
	Duration  int
	Size      int64
	FileName  string
}

type UpdatePostParams struct {
	UserID     int64
	PostID     int64
	Content    *string
	Visibility *PostVisibility
}
