package domain

import (
	"context"
	"encoding/json"
	"time"
)

const (
	EventEmailRegister = "email.register"
)

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
	Close()
}

type EventPayload struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}
