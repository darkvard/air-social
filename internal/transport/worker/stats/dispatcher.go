package stats

import (
	"context"
	"fmt"
	"strconv"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

const (
	SystemName          = "social"
	FeatureStats        = "stats"
	StatePostLikes      = "post_likes"
	StatePostShares     = "post_shares"
	StatePostComments   = "post_comments"
	StateCommentLikes   = "comment_likes"
	StateCommentReplies = "comment_replies"
)

type Dispatcher struct {
	cache    common.Cache
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(cache common.Cache) *Dispatcher {
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

func (d *Dispatcher) updateCache(ctx context.Context, state string, id int64, incr int64) error {
	key := common.BuildCacheKey(SystemName, FeatureStats, state, "")
	field := strconv.FormatInt(id, 10)
	_, err := d.cache.HIncrBy(ctx, key, field, incr)
	return err
}

func (d *Dispatcher) handlePostLike(ctx context.Context, p common.LikeEventPayload) error {
	incr := int64(-1)
	if p.IsLiked {
		incr = 1
	}
	return d.updateCache(ctx, StatePostLikes, p.TargetID, incr)
}

func (d *Dispatcher) handleCommentCreated(ctx context.Context, p common.CommentEventPayload) error {
	if p.ParentID == nil {
		return d.updateCache(ctx, StatePostComments, p.PostID, 1)
	}
	return d.updateCache(ctx, StateCommentReplies, *p.ParentID, 1)
}

func (d *Dispatcher) handleCommentDeleted(ctx context.Context, p common.CommentEventPayload) error {
	if p.ParentID == nil {
		return d.updateCache(ctx, StatePostComments, p.PostID, -1)
	}
	return d.updateCache(ctx, StateCommentReplies, *p.ParentID, -1)
}

func (d *Dispatcher) handlePostShare(ctx context.Context, p common.ShareEventPayload) error {
	incr := int64(-1)
	if p.IsShared {
		incr = 1
	}
	return d.updateCache(ctx, StatePostShares, p.OriginalPostID, incr)
}

func (d *Dispatcher) handleCommentLike(ctx context.Context, p common.LikeEventPayload) error {
	incr := int64(-1)
	if p.IsLiked {
		incr = 1
	}
	return d.updateCache(ctx, StateCommentLikes, p.TargetID, incr)
}
