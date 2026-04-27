package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/pkg"
)

type QueryDeps struct {
	ConvRepo chat.ConversationRepository
	MsgRepo  chat.MessageRepository
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
		return nil, pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil, pkg.ErrNotFound
	}

	if conv.FindParticipant(userID) == nil {
		return nil, pkg.ErrForbidden
	}

	count, err := u.deps.Unread.Get(ctx, userID, convID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	conv.UnreadCount = int(count)

	u.populateLastMessage(ctx, conv)

	return conv, nil
}

func (u *ConversationQueryUseCase) GetConversations(ctx context.Context, params chat.GetConversationsParams) (common.CursorPaginatedResult[chat.Conversation, string], error) {
	var empty common.CursorPaginatedResult[chat.Conversation, string]

	if params.State == "" {
		params.State = chat.StateActive
	}
	params.Query.NormalizePagination()

	convs, err := u.deps.ConvRepo.GetList(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	unreadMap, err := u.deps.Unread.GetAll(ctx, params.UserID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}
	for i := range convs {
		convs[i].UnreadCount = int(unreadMap[convs[i].ID])
	}

	u.populateLastMessages(ctx, convs)

	return common.NewCursorPaginatedResult(convs, params.Query.Limit), nil
}

// populateLastMessage fetches the last message for a single conversation.
// Errors are swallowed — LastMessage stays nil if the fetch fails.
func (u *ConversationQueryUseCase) populateLastMessage(ctx context.Context, conv *chat.Conversation) {
	if conv.LastMsgID == "" {
		return
	}
	msgs, err := u.deps.MsgRepo.GetByIDs(ctx, []string{conv.LastMsgID})
	if err != nil || len(msgs) == 0 {
		return
	}
	conv.LastMessage = &msgs[0]
}

// populateLastMessages bulk-fetches the last message for a list of conversations in one MongoDB round-trip.
// Errors are swallowed — individual LastMessage fields stay nil if the fetch fails.
func (u *ConversationQueryUseCase) populateLastMessages(ctx context.Context, convs []chat.Conversation) {
	ids := make([]string, 0, len(convs))
	for _, c := range convs {
		if c.LastMsgID != "" {
			ids = append(ids, c.LastMsgID)
		}
	}
	if len(ids) == 0 {
		return
	}

	msgs, err := u.deps.MsgRepo.GetByIDs(ctx, ids)
	if err != nil {
		return
	}

	byID := make(map[string]*chat.Message, len(msgs))
	for i := range msgs {
		byID[msgs[i].ID] = &msgs[i]
	}
	for i := range convs {
		if m, ok := byID[convs[i].LastMsgID]; ok {
			convs[i].LastMessage = m
		}
	}
}
