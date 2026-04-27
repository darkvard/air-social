package usecase

import (
	"context"

	"air-social/internal/domain/chat"
)

type noopRTPublisher struct{}

func (noopRTPublisher) PublishNewMessage(_ context.Context, _ *chat.Message, _ *chat.Conversation) error {
	return nil
}

func (noopRTPublisher) PublishReadEvent(_ context.Context, _ string, _ int64, _ string) error {
	return nil
}

func (noopRTPublisher) PublishMessageDeleted(_ context.Context, _ string, _ string) error {
	return nil
}

func (noopRTPublisher) PublishReactionAdded(_ context.Context, _ string, _ string, _ chat.Reaction) error {
	return nil
}

func (noopRTPublisher) PublishReactionRemoved(_ context.Context, _ string, _ string, _ int64) error {
	return nil
}

func NewNoopRTPublisher() RealtimePublisher { return noopRTPublisher{} }
