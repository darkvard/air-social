package mq

import (
	"context"
	"errors"

	"air-social/internal/domain"
)

type eventHandlerImpl struct {
	callCount  map[string]int
	maxAttempt int
}

func newEventHandler() *eventHandlerImpl {
	return &eventHandlerImpl{
		callCount:  make(map[string]int),
		maxAttempt: 3,
	}
}

func (e *eventHandlerImpl) Handle(ctx context.Context, evt domain.EventPayload) error {
	e.callCount[evt.EventType]++
	attempt := e.callCount[evt.EventType]

	logInfo(consumer, "handle event", "Type = %s", evt.EventType)

	if attempt > e.maxAttempt {
		return nil //ack
	}

	if evt.EventType == emailFailKey {
		logError(consumer, "Simulate error", "Type = %s, Attempt = %d", evt.EventType, attempt)
		return errors.New("force retry")
	}

	return nil
}
