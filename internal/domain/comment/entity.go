package comment

import (
	"time"

	"air-social/internal/domain/stats"
)

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	ParentID  *int64
	Content   string
	Media     []Media
	Version   int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Stat    stats.CommentStats
	IsLiked *bool
	Author  *Author
}

func (c Comment) GetCursor() int64 {
	return c.ID
}

type Media struct {
	MediaKey  string
	MediaType string
}

type Author struct {
	ID         int64
	Username   string
	Avatar     string
	IsVerified bool
}
