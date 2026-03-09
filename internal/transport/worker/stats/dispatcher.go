package stats

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/internal/domain/stats/cache"
	"air-social/pkg"
)

type Dispatcher struct {
	cache    cache.Provider
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(cache cache.Provider) *Dispatcher {
	disp := &Dispatcher{
		cache:    cache,
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
	d.handlers[common.EventPostLike] = makeHandler(d.handlePostLike)
	d.handlers[common.EventPostShare] = makeHandler(d.handlePostShare)
	d.handlers[common.EventCommentCreated] = makeHandler(d.handleCommentCreated)
	d.handlers[common.EventCommentDeleted] = makeHandler(d.handleCommentDeleted)
	d.handlers[common.EventCommentLike] = makeHandler(d.handleCommentLike)
}

func makeHandler[T any](handlerFunc func(ctx context.Context, payload T) error) common.EventHandler {
	return func(ctx context.Context, event common.Event) error {
		payload, err := common.UnmarshalEvent[T](event.Data)
		if err != nil {
			return fmt.Errorf("stats dispatcher unmarshal payload error: %w", err)
		}
		return handlerFunc(ctx, payload)
	}
}

func (d *Dispatcher) handlePostLike(ctx context.Context, p common.LikeEventPayload) error {
	incr := int64(-1)
	if p.IsLiked {
		incr = 1
	}
	return d.cache.UpdateStatsHash(ctx, cache.StatePostLikes, p.TargetID, incr)
}

func (d *Dispatcher) handleCommentCreated(ctx context.Context, p common.CommentEventPayload) error {
	if p.ParentID == nil {
		return d.cache.UpdateStatsHash(ctx, cache.StatePostComments, p.PostID, 1)
	}
	return d.cache.UpdateStatsHash(ctx, cache.StateCommentReplies, *p.ParentID, 1)
}

func (d *Dispatcher) handleCommentDeleted(ctx context.Context, p common.CommentEventPayload) error {
	if p.ParentID == nil {
		return d.cache.UpdateStatsHash(ctx, cache.StatePostComments, p.PostID, -1)
	}
	return d.cache.UpdateStatsHash(ctx, cache.StateCommentReplies, *p.ParentID, -1)
}

func (d *Dispatcher) handlePostShare(ctx context.Context, p common.ShareEventPayload) error {
	incr := int64(-1)
	if p.IsShared {
		incr = 1
	}
	return d.cache.UpdateStatsHash(ctx, cache.StatePostShares, p.OriginalPostID, incr)
}

func (d *Dispatcher) handleCommentLike(ctx context.Context, p common.LikeEventPayload) error {
	incr := int64(-1)
	if p.IsLiked {
		incr = 1
	}
	return d.cache.UpdateStatsHash(ctx, cache.StateCommentLikes, p.TargetID, incr)
}
