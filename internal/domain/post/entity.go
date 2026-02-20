package post

import (
	"time"
)

const (
	VisibilityPublic    Visibility = "public"
	VisibilityFollowers Visibility = "followers"
	VisibilityPrivate   Visibility = "private"
)

type Visibility string

type Post struct {
	ID         int64
	UserID     int64
	Content    string
	Visibility Visibility
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time

	IsLiked        bool
	OriginalPostID *int64
	Counts         Counts
	Media          []Media
	Author         *Author
}

func (p Post) GetCursor() int64 {
	return p.ID
}

type Author struct {
	ID         int64
	FullName   string
	Avatar     string
	IsVerified bool
}

type Counts struct {
	LikesCount    int
	CommentsCount int
	SharesCount   int
}

type Media struct {
	ID        int64
	PostID    int64
	MediaKey  string
	MediaType string
	Metadata  MediaMetadata
	CreatedAt time.Time
}

type MediaMetadata struct {
	Width    int
	Height   int
	Duration int
	Size     int64
	FileName string
}
