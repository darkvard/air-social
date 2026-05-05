package notification

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	notidomain "air-social/internal/domain/notification"
	"air-social/pkg"
)

type Dispatcher struct {
	notifUC  notidomain.UseCase
	handlers map[common.EventType]common.EventHandler
}

func NewDispatcher(notifUC notidomain.UseCase) *Dispatcher {
	d := &Dispatcher{
		notifUC:  notifUC,
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
	d.handlers[common.EventPostLike] = makeHandler(d.handlePostLike)
	d.handlers[common.EventCommentLike] = makeHandler(d.handleCommentLike)
	d.handlers[common.EventCommentCreated] = makeHandler(d.handleCommentCreated)
	d.handlers[common.EventPostShare] = makeHandler(d.handlePostShare)
	d.handlers[common.EventFollowCreated] = makeHandler(d.handleFollowCreated)
	d.handlers[common.EventMessageCreated] = makeHandler(d.handleMessageCreated)
}

func makeHandler[T any](fn func(ctx context.Context, payload T) error) common.EventHandler {
	return func(ctx context.Context, event common.Event) error {
		payload, err := common.UnmarshalEvent[T](event.Data)
		if err != nil {
			return fmt.Errorf("notification dispatcher unmarshal: %w", err)
		}
		return fn(ctx, payload)
	}
}

func (d *Dispatcher) handlePostLike(ctx context.Context, p common.LikeEventPayload) error {
	if p.ActorID == p.OwnerID {
		return nil
	}
	if !p.IsLiked {
		return d.notifUC.Delete(ctx, p.OwnerID, p.ActorID, notidomain.NotifPostLiked, &p.TargetID)
	}
	return d.notifUC.Create(ctx, &notidomain.Notification{
		UserID:     p.OwnerID,
		ActorID:    p.ActorID,
		Type:       notidomain.NotifPostLiked,
		TargetID:   &p.TargetID,
		TargetType: "post",
	})
}

func (d *Dispatcher) handleCommentLike(ctx context.Context, p common.LikeEventPayload) error {
	if p.ActorID == p.OwnerID {
		return nil
	}
	if !p.IsLiked {
		return d.notifUC.Delete(ctx, p.OwnerID, p.ActorID, notidomain.NotifCommentLiked, &p.TargetID)
	}
	return d.notifUC.Create(ctx, &notidomain.Notification{
		UserID:     p.OwnerID,
		ActorID:    p.ActorID,
		Type:       notidomain.NotifCommentLiked,
		TargetID:   &p.TargetID,
		TargetType: "comment",
	})
}

func (d *Dispatcher) handleCommentCreated(ctx context.Context, p common.CommentEventPayload) error {
	if p.PostOwnerID == 0 || p.ActorID == p.PostOwnerID {
		return nil
	}
	return d.notifUC.Create(ctx, &notidomain.Notification{
		UserID:     p.PostOwnerID,
		ActorID:    p.ActorID,
		Type:       notidomain.NotifCommentCreated,
		TargetID:   &p.PostID,
		TargetType: "post",
	})
}

func (d *Dispatcher) handlePostShare(ctx context.Context, p common.ShareEventPayload) error {
	if p.ActorID == p.OwnerID {
		return nil
	}
	if !p.IsShared {
		return d.notifUC.Delete(ctx, p.OwnerID, p.ActorID, notidomain.NotifPostShared, &p.OriginalPostID)
	}
	return d.notifUC.Create(ctx, &notidomain.Notification{
		UserID:     p.OwnerID,
		ActorID:    p.ActorID,
		Type:       notidomain.NotifPostShared,
		TargetID:   &p.OriginalPostID,
		TargetType: "post",
	})
}

func (d *Dispatcher) handleFollowCreated(ctx context.Context, p common.FollowCreatedPayload) error {
	return d.notifUC.Create(ctx, &notidomain.Notification{
		UserID:  p.FolloweeID,
		ActorID: p.FollowerID,
		Type:    notidomain.NotifUserFollowed,
	})
}

func (d *Dispatcher) handleMessageCreated(ctx context.Context, p common.MessageEventPayload) error {
	for _, recipientID := range p.RecipientIDs {
		_ = d.notifUC.Create(ctx, &notidomain.Notification{
			UserID:     recipientID,
			ActorID:    p.SenderID,
			Type:       notidomain.NotifMessageRequest,
			TargetType: "conversation",
			Data:       map[string]any{"conversation_id": p.ConversationID},
		})
	}
	return nil
}
