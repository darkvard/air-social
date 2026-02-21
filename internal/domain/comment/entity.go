package comment

import (
	"time"
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

	IsLiked bool
	Stat    Stat
}

type Media struct {
	URL  string
	Type string
}

type Stat struct {
	CommentID    int64
	LikesCount   int32
	RepliesCount int32
	UpdatedAt    time.Time
}
