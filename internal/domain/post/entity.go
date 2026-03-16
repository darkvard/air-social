package post

import (
	"time"

	"air-social/internal/domain/stats"
)

const (
	VisibilityPublic    Visibility = "public"
	VisibilityFollowers Visibility = "followers"
	VisibilityPrivate   Visibility = "private"
)

type Visibility string

type Post struct {
	ID             int64
	UserID         int64
	Content        string
	Visibility     Visibility
	Version        int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	OriginalPostID *int64

	Media   []Media
	Stat    stats.PostStats
	IsLiked *bool
	Author  *Author
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

type Media struct {
	ID        int64
	PostID    int64
	MediaKey  string
	MediaType string
	Metadata  MediaMetadata
	CreatedAt time.Time
}

type MediaMetadata struct {
	Width    int32
	Height   int32
	Duration int32
	Size     int64
	FileName string
}

type Sharer struct {
	ShareID    int64 // This is the Post ID of the share, used for cursor
	UserID     int64
	FullName   string
	Avatar     string
	IsVerified bool
}

func (s Sharer) GetCursor() int64 {
	return s.ShareID
}
