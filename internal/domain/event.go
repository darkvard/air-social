package domain

import (
	"context"
	"time"
)

type EventQueue interface {
	Publish(ctx context.Context, topic string, payload any) error
	Close()
}

type EventPayload struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}
