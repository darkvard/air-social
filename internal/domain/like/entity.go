package like

import "time"

type Action string

const (
	ActionLiked   Action = "LIKED"
	ActionUnliked Action = "UNLIKED"
)

type PostLike struct {
	PostID    int64
	UserID    int64
	CreatedAt time.Time
}

type CommentLike struct {
	CommentID int64
	UserID    int64
	CreatedAt time.Time
}
