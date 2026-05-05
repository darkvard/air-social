package notification

import "time"

type NotificationType string

const (
	NotifPostLiked      NotificationType = "post_liked"
	NotifCommentLiked   NotificationType = "comment_liked"
	NotifCommentCreated NotificationType = "comment_created"
	NotifPostShared     NotificationType = "post_shared"
	NotifUserFollowed   NotificationType = "user_followed"
	NotifMessageRequest NotificationType = "message_request"
)

type Notification struct {
	ID         int64
	UserID     int64
	ActorID    int64
	Type       NotificationType
	TargetID   *int64
	TargetType string
	Data       map[string]any
	Read       bool
	ReadAt     *time.Time
	CreatedAt  time.Time
}

func (n Notification) GetCursor() int64 { return n.ID }
