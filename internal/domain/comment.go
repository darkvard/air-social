package domain

import (
	"context"
	"time"
)
// todo: remove

type CommentRepository interface {
	Create(ctx context.Context, comment *Comment) error
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Comment, error)
	GetByPostID(ctx context.Context, postID int64, cursor int64, limit int) ([]Comment, error)
	GetReplies(ctx context.Context, parentID int64, cursor int64, limit int) ([]Comment, error)
}

type Comment struct {
	ID         int64
	PostID     int64
	UserID     int64
	ParentID   *int64
	Content    string
	IsLiked    bool
	LikesCount int
	Media      []CommentMedia

	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CommentMedia struct {
	URL  string
	Type string
}
