package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"air-social/pkg"
)

const (
	EventEmailVerify        EventType = EventType("email.verify")
	EventEmailResetPassword EventType = EventType("email.reset_password")
)

const (
	EventPostCreated EventType = "social.post.created"
	EventPostDeleted EventType = "social.post.deleted"
	EventPostLike    EventType = "social.post.like"
	EventPostShare   EventType = "social.post.share"

	EventCommentLike    EventType = "social.comment.like"
	EventCommentCreated EventType = "social.comment.created"
	EventCommentDeleted EventType = "social.comment.deleted"
)

type EventType string

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event Event) error
}

type EventHandler func(ctx context.Context, event Event) error

type Event struct {
	ID        string
	Typ       EventType
	Timestamp time.Time
	Data      any
}

func NewEvent(typ EventType, data any) Event {
	return Event{
		ID:        uuid.NewString(),
		Typ:       typ,
		Timestamp: pkg.TimeNowUTC(),
		Data:      data,
	}
}

func UnmarshalEvent[T any](data any) (T, error) {
	var target T
	bytes, err := json.Marshal(data)
	if err != nil {
		return target, err
	}
	if err := json.Unmarshal(bytes, &target); err != nil {
		return target, err
	}
	return target, nil
}

type EmailEventPayload struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Expiry string `json:"expiry"`
}

type LikeEventPayload struct {
	TargetID int64     `json:"target_id"`
	ActorID  int64     `json:"actor_id"`
	OwnerID  int64     `json:"owner_id"`
	IsLiked  bool      `json:"is_liked"`
	Typ      EventType `json:"type"`
}

type CommentEventPayload struct {
	PostID    int64     `json:"post_id"`
	CommentID int64     `json:"comment_id"`
	ParentID  *int64    `json:"parent_id,omitempty"` // nil ? root comment : reply comment
	ActorID   int64     `json:"actor_id"`
	OwnerID   int64     `json:"owner_id"`
	Typ       EventType `json:"type"`
}

func (c CommentEventPayload) IsReply() bool {
	original := c.ParentID
	return original != nil && *original > 0
}

type ShareEventPayload struct {
	OriginalPostID int64 `json:"original_post_id"`
	NewPostID      int64 `json:"new_post_id"`
	ActorID        int64 `json:"actor_id"`
	IsShared       bool  `json:"is_shared"` // true ? create post (original_id != nil) : delete post (original_id != nil)
}

type PostFeedEventPayload struct {
	PostID    int64 `json:"post_id"`
	AuthorID  int64 `json:"author_id"`
	Timestamp int64 `json:"timestamp"` // UnixMilli, used as score in Redis sorted set
}
