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

// FromDomainEvent converts a domain event to an infrastructure model for JSON marshaling.
func FromDomainEvent(e domain.Event) Event {
	// Map specific data types if needed to ensure correct JSON tags
	var data any = e.Data
	if d, ok := e.Data.(domain.EmailEvent); ok {
		data = EventEmailData{
			Email:  d.Email,
			Name:   d.Name,
			Link:   d.Link,
			Expiry: d.Expiry,
		}
	}

	return Event{
		EventID:   e.EventID,
		EventType: string(e.EventType),
		Timestamp: e.Timestamp,
		Data:      data,
	}
}

func (m *Event) ToDomainEvent() domain.Event {
	var data any = m.Data

	// Attempt to convert map[string]interface{} (from JSON unmarshal) to specific struct if needed.
	// Note: In a generic consumer, Data is often kept as map[string]any until the specific handler processes it.
	// However, if we know the structure, we can map it here.
	// For now, we keep it as is or let the Service layer handle the map decoding.

	return domain.Event{
		EventID:   m.EventID,
		EventType: domain.EventType(m.EventType),
		Timestamp: m.Timestamp,
		Data:      data,
	}
}
