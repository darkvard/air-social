package messages

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
)

// Declared here because MessageUseCase is the consumer of this dependency.

type realtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

type MessageUseCase struct {
	msgRepo     chat.MessageRepository
	convRepo    chat.ConversationRepository
	event       common.EventPublisher
	rtPublisher realtimePublisher
	unread      chat.UnreadStore
}

func NewMessageUseCase(
	msgRepo chat.MessageRepository,
	convRepo chat.ConversationRepository,
	event common.EventPublisher,
	rtPublisher realtimePublisher,
	unread chat.UnreadStore,
) *MessageUseCase {
	return &MessageUseCase{
		msgRepo:     msgRepo,
		convRepo:    convRepo,
		event:       event,
		rtPublisher: rtPublisher,
		unread:      unread,
	}
}

func (u *MessageUseCase) GetMessages(ctx context.Context, params chat.GetMessagesParams) (common.CursorPaginatedResult[chat.Message, string], error) {
	var empty common.CursorPaginatedResult[chat.Message, string]
	return empty, nil
}

func (u *MessageUseCase) SendMessage(ctx context.Context, params chat.SendMessageParams) (*chat.Message, error) {
	return nil, nil
}

func (u *MessageUseCase) EditMessage(ctx context.Context, params chat.EditMessageParams) (*chat.Message, error) {
	return nil, nil
}

func (u *MessageUseCase) DeleteMessage(ctx context.Context, msgID string, userID int64) error {
	return nil
}

func (u *MessageUseCase) AddReaction(ctx context.Context, msgID string, userID int64, emoji string) error {
	return nil
}

func (u *MessageUseCase) RemoveReaction(ctx context.Context, msgID string, userID int64) error {
	return nil
}

func (u *MessageUseCase) MarkRead(ctx context.Context, convID string, userID int64, lastReadMsgID string) error {
	return nil
}
