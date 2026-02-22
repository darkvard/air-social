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
	ID             int64
	UserID         int64
	Content        string
	Visibility     Visibility
	OriginalPostID *int64
	Version        int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time

	IsLiked bool
	Media   []Media
	Stat    Stat
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

type Stat struct {
	PostID        int64
	LikesCount    int32
	CommentsCount int32
	SharesCount   int32
	UpdatedAt     time.Time
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
