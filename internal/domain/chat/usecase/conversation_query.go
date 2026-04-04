package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
)

type QueryDeps struct {
	ConvRepo chat.ConversationRepository
	Unread   chat.UnreadStore
}

type ConversationQueryUseCase struct {
	deps QueryDeps
}

func NewQueryUseCase(d QueryDeps) *ConversationQueryUseCase {
	return &ConversationQueryUseCase{deps: d}
}

func (u *ConversationQueryUseCase) GetConversations(ctx context.Context, params chat.GetConversationsParams) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}

func (u *ConversationQueryUseCase) GetPendingConversations(ctx context.Context, userID int64, q common.CursorQueryParams[string]) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}

func (u *ConversationQueryUseCase) GetConversation(ctx context.Context, convID string, userID int64) (*chat.Conversation, error) {
	return nil, nil
}
