package domain

import (
	"context"
	"time"
)

const (
	EmailVerify        = "email.verify"
	EmailResetPassword = "email.reset.password"
)

type EventHandler interface {
	Handle(ctx context.Context, evt EventPayload) error
}

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
	Close()
}

type EventPayload struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

type EventEmailData struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Link   string `json:"link"`
	Expiry string `json:"expiry"`
}
