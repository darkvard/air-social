package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
)

type MessageDeps struct {
	MsgRepo     chat.MessageRepository
	ConvRepo    chat.ConversationRepository
	Event       common.EventPublisher
	RTPublisher RealtimePublisher
	Unread      chat.UnreadStore
}

type MessageUseCase struct {
	deps MessageDeps
}

func NewMessageUseCase(d MessageDeps) *MessageUseCase {
	return &MessageUseCase{deps: d}
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

func (u *MessageUseCase) AddReaction(ctx context.Context, msgID string, userID int64, reaction chat.ReactionType) error {
	return nil
}

func (u *MessageUseCase) RemoveReaction(ctx context.Context, msgID string, userID int64) error {
	return nil
}

func (u *MessageUseCase) MarkRead(ctx context.Context, convID string, userID int64, lastReadMsgID string) error {
	return nil
}
