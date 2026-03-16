package like

import (
	"time"

	"air-social/internal/domain/common"
)

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

type Liker struct {
	LikeID     int64 // post_id or comment_id
	UserID     int64
	FullName   string
	Avatar     string
	IsVerified bool
}

func (l Liker) GetCursor() int64 {
	return l.LikeID
}

type GetCursorParams struct {
	TargetID int64
	Query    common.CursorQueryParams[int64]
}
