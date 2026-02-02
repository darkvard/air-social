package domain

import (
	"context"
	"time"
)

type EventType string

const (
	EmailVerify        EventType = "email.verify"
	EmailResetPassword EventType = "email.reset.password"
)

type EventDispatcher interface {
	Dispatch(ctx context.Context, evt Event) error
}

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
	Close()
}

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
