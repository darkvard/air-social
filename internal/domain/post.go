package domain

import (
	"context"
	"time"
)
// todo: remove

const (
	VisibilityPublic    PostVisibility = "public"
	VisibilityFollowers PostVisibility = "followers"
	VisibilityPrivate   PostVisibility = "private"
)

type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id int64) error
	IsOwner(ctx context.Context, postID, userID int64) (bool, error)
	GetByID(ctx context.Context, postID, userID int64) (*Post, error)
	GetUserPosts(ctx context.Context, userID int64, params CursorQueryParams) ([]Post, error)
}

type PostVisibility string

type Post struct {
	ID             int64
	UserID         int64
	Content        string
	Visibility     PostVisibility
	OriginalPostID *int64
	IsLiked        bool

	Counts PostCounts
	Media  []PostMedia
	User   *UserSummary

	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type PostCounts struct {
	LikesCount    int
	CommentsCount int
	SharesCount   int
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
