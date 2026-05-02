package ws

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"air-social/internal/domain/chat"
)

const (
	keyUserID        = "user_id"
	keyMessageID     = "message_id"
	keyLastReadMsgID = "last_read_msg_id"
	keyReaction      = "reaction"
)

type HubPublisher struct {
	redis *redis.Client
}

func NewHubPublisher(redis *redis.Client) *HubPublisher {
	return &HubPublisher{
		redis: redis,
	}
}

func (p *HubPublisher) publish(ctx context.Context, convID string, event OutboundEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.redis.Publish(ctx, chatChannel(convID), payload).Err()
}

func (p *HubPublisher) PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error {
	return p.publish(ctx, conv.ID, OutboundEvent{Type: EventNewMessage, Data: msg})
}

func (p *HubPublisher) PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error {
	return p.publish(ctx, convID, OutboundEvent{
		Type: EventRead,
		Data: map[string]any{keyUserID: userID, keyLastReadMsgID: lastReadMsgID},
	})
}

func (p *HubPublisher) PublishMessageDeleted(ctx context.Context, msgID string, convID string) error {
	return p.publish(ctx, convID, OutboundEvent{
		Type: EventMessageDeleted,
		Data: map[string]any{keyMessageID: msgID},
	})
}

func (p *HubPublisher) PublishReactionAdded(ctx context.Context, msgID string, convID string, reaction chat.Reaction) error {
	return p.publish(ctx, convID, OutboundEvent{
		Type: EventReactionAdded,
		Data: map[string]any{keyMessageID: msgID, keyReaction: reaction},
	})
}

func (p *HubPublisher) PublishReactionRemoved(ctx context.Context, msgID string, convID string, userID int64) error {
	return p.publish(ctx, convID, OutboundEvent{
		Type: EventReactionRemoved,
		Data: map[string]any{keyMessageID: msgID, keyUserID: userID},
	})
}
