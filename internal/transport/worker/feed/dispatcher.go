package feed

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/internal/domain/feed"
	"air-social/pkg"
)

type Dispatcher struct {
	feedUC   feed.UseCase
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(feedUC feed.UseCase) *Dispatcher {
	disp := &Dispatcher{
		feedUC:   feedUC,
		handlers: make(map[common.EventType]common.EventHandler),
	}
	disp.registerHandlers()
	return disp
}

func (d *Dispatcher) Dispatch(ctx context.Context, event common.Event) error {
	handler, ok := d.handlers[event.Typ]
	if !ok {
		return fmt.Errorf("no handler for event type %s: %w", event.Typ, pkg.ErrNoEventHandler)
	}
	return handler(ctx, event)
}

func (d *Dispatcher) registerHandlers() {
	d.handlers[common.EventPostCreated] = makeHandler(d.handlePostCreated)
	d.handlers[common.EventPostDeleted] = makeHandler(d.handlePostDeleted)
}

func makeHandler[T any](handlerFunc func(ctx context.Context, payload T) error) common.EventHandler {
	return func(ctx context.Context, event common.Event) error {
		payload, err := common.UnmarshalEvent[T](event.Data)
		if err != nil {
			return fmt.Errorf("feed dispatcher unmarshal payload error: %w", err)
		}
		return handlerFunc(ctx, payload)
	}
}

func (d *Dispatcher) handlePostCreated(ctx context.Context, p common.PostFeedEventPayload) error {
	return d.feedUC.Command.DistributePost(ctx, p.PostID, p.AuthorID, p.Timestamp)
}

func (d *Dispatcher) handlePostDeleted(ctx context.Context, p common.PostFeedEventPayload) error {
	return d.feedUC.Command.RevokePost(ctx, p.PostID, p.AuthorID)
}
