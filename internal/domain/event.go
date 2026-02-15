package domain

import (
	"context"
	"time"
)

type EventType string

const (
	// routing keys (service uses these to publish)
	RoutingKeyEmailVerify        = "email.verify"
	RoutingKeyEmailResetPassword = "email.reset_password"

	// event types (worker uses these to dispatch)
	EmailVerify        EventType = EventType(RoutingKeyEmailVerify)
	EmailResetPassword EventType = EventType(RoutingKeyEmailResetPassword)
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, evt Event) error
}

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
	Close()
}

type EventHandler func(ctx context.Context, evt Event) error

type Event struct {
	EventID   string
	EventType EventType
	Timestamp time.Time
	Data      any
}

type EmailEvent struct {
	Email  string
	Name   string
	Link   string
	Expiry string
}
