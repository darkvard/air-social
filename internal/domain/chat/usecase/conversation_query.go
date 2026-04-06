package usecase

import (
	"context"
	"slices"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/pkg"
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

func (u *ConversationQueryUseCase) GetConversation(ctx context.Context, convID string, userID int64) (*chat.Conversation, error) {
	conv, err := u.deps.ConvRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	// author
	isMember := slices.ContainsFunc(conv.Participants, func(p chat.Participant) bool {
		return p.UserID == userID
	})
	if !isMember {
		return nil, pkg.ErrUnauthorized
	}

	// unread count — Unread may be nil until UnreadStore is implemented
	if u.deps.Unread != nil {
		count, err := u.deps.Unread.Get(ctx, userID, convID)
		if err != nil {
			return nil, pkg.OrInternalError(err)
		}
		conv.UnreadCount = int(count)
	}

	// TODO: populate conv.LastMessage from MessageRepository once message layer is implemented.

	return conv, nil
}

func (u *ConversationQueryUseCase) GetConversations(ctx context.Context, params chat.GetConversationsParams) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}

func (u *ConversationQueryUseCase) GetPendingConversations(ctx context.Context, userID int64, q common.CursorQueryParams[string]) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]
	return empty, nil
}
