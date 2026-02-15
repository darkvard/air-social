package model

import (
	"time"

	"air-social/internal/domain"
)

type Event struct {
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

func FromDomainEvent(e domain.Event) Event {
	return Event{
		EventID:   e.EventID,
		EventType: string(e.EventType),
		Timestamp: e.Timestamp,
		Data:      e.Data,
	}
}

func (m *Event) ToDomainEvent() domain.Event {
	return domain.Event{
		EventID:   m.EventID,
		EventType: domain.EventType(m.EventType),
		Timestamp: m.Timestamp,
		Data:      m.Data,
	}
}
