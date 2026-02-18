package shared

import (
	"context"
	"time"
)

const (
	EventVerify        EventType = EventType("verify")
	EventResetPassword EventType = EventType("reset_password")
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, evt Event) error
}

// todo: ben infra ko can routing key nua, data tuong minh, roi lay routing key tu event type
type EventPublisher interface {
	Publish(ctx context.Context, evt Event) error
}

type EventType string

type Event struct {
	ID        string
	Typ       EventType
	Timestamp time.Time
	Data      any
}

type EmailPayload struct {
	Email  string
	Name   string
	Link   string
	Expiry string
}
