package conversation

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
)

type QueryUseCase struct {
	convRepo chat.ConversationRepository
	unread   chat.UnreadStore
}

func NewQueryUseCase(convRepo chat.ConversationRepository, unread chat.UnreadStore) *QueryUseCase {
	return &QueryUseCase{convRepo: convRepo, unread: unread}
}

func (u *QueryUseCase) GetConversations(ctx context.Context, params chat.GetConversationsParams) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}

func (u *QueryUseCase) GetPendingConversations(ctx context.Context, userID int64, q common.CursorQueryParams[string]) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}

func (u *QueryUseCase) GetConversation(ctx context.Context, convID string, userID int64) (*chat.Conversation, error) {
	return nil, nil
}
