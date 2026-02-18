package shared

import (
	"context"
	"encoding/json"
	"time"
)

const (
	EventVerify        EventType = EventType("verify")
	EventResetPassword EventType = EventType("reset_password")
)

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, event Event) error
}

type EventHandler func(ctx context.Context, event Event) error

type EventType string

type Event struct {
	ID        string
	Typ       EventType
	Timestamp time.Time
	Data      any
}

type EmailEventPayload struct {
	Email  string
	Name   string
	Link   string
	Expiry string
}

type NotificationEventPayload struct {
	
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
