package mq

import (
	"context"
	"errors"

	"air-social/internal/domain"
)

type eventHandlerImpl struct{}

func (e *eventHandlerImpl) Handle(ctx context.Context, evt domain.EventPayload) error {
	logInfo(consumer, "RECEIVED EVENT", "Type: %s", evt.EventType)

	switch evt.EventType {
	case emailFailKey:
		logError(consumer, "SIMULATE ERROR", "Type: %s (NACK)", evt.EventType)
		return errors.New("force nack")
	default:
		logInfo(consumer, "HANDLE SUCCESS", "Type: %s (ACK)", evt.EventType)
		return nil
	}
}
