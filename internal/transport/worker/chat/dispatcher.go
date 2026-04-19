package chat

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type ConvAutoAcceptor interface {
	AcceptPendingByFollowEvent(ctx context.Context, followerID, followeeID int64) error
}

type Dispatcher struct {
	acceptor ConvAutoAcceptor
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(acceptor ConvAutoAcceptor) *Dispatcher {
	d := &Dispatcher{
		acceptor: acceptor,
		handlers: make(map[common.EventType]common.EventHandler),
	}
	d.registerHandlers()
	return d
}

func (d *Dispatcher) Dispatch(ctx context.Context, event common.Event) error {
	handler, ok := d.handlers[event.Typ]
	if !ok {
		return fmt.Errorf("no handler for event type %s: %w", event.Typ, pkg.ErrNoEventHandler)
	}
	return handler(ctx, event)
}

func (d *Dispatcher) registerHandlers() {
	d.handlers[common.EventFollowCreated] = makeHandler(d.handleFollowCreated)
}

func makeHandler[T any](fn func(ctx context.Context, payload T) error) common.EventHandler {
	return func(ctx context.Context, event common.Event) error {
		payload, err := common.UnmarshalEvent[T](event.Data)
		if err != nil {
			return fmt.Errorf("chat dispatcher unmarshal error: %w", err)
		}
		return fn(ctx, payload)
	}
}

func (d *Dispatcher) handleFollowCreated(ctx context.Context, p common.FollowCreatedPayload) error {
	return d.acceptor.AcceptPendingByFollowEvent(ctx, p.FollowerID, p.FolloweeID)
}
