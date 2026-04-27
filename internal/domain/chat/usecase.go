package chat

import (
	"context"

	"air-social/internal/domain/common"
)

type ConversationQuerier interface {
	GetConversations(ctx context.Context, params GetConversationsParams) (common.CursorPaginatedResult[Conversation, string], error)
	GetConversation(ctx context.Context, convID string, userID int64) (*Conversation, error)
}

type ConversationWriter interface {
	CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*Conversation, bool, error)
	CreateGroup(ctx context.Context, params CreateGroupParams) (*Conversation, bool, error)
	UpdateGroup(ctx context.Context, params UpdateGroupParams) (*Conversation, error)
}

type ConversationMember interface {
	AcceptConversation(ctx context.Context, convID string, userID int64) error
	AcceptPendingByFollowEvent(ctx context.Context, followerID, followeeID int64) error
	AddMember(ctx context.Context, p AddMemberParams) error
	RemoveMember(ctx context.Context, p RemoveMemberParams) error
	UpdateMemberRole(ctx context.Context, p UpdateMemberRoleParams) error
}

type Messenger interface {
	GetMessages(ctx context.Context, params GetMessagesParams) (common.CursorPaginatedResult[Message, string], error)
	SendMessage(ctx context.Context, params SendMessageParams) (*Message, bool, error)
	EditMessage(ctx context.Context, params EditMessageParams) (*Message, error)
	DeleteMessage(ctx context.Context, msgID string, userID int64) error
	AddReaction(ctx context.Context, msgID string, userID int64, reaction ReactionType) error
	RemoveReaction(ctx context.Context, msgID string, userID int64) error
	MarkRead(ctx context.Context, convID string, userID int64) error
	MarkUnread(ctx context.Context, convID string, userID int64) error
}

type ConversationUseCase struct {
	Query  ConversationQuerier
	Write  ConversationWriter
	Member ConversationMember
}

type UseCase struct {
	Conversation ConversationUseCase
	Messages     Messenger
}
