package chat

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
)

type ConversationsUseCase interface {
	GetConversations(ctx context.Context, params GetConversationsParams) (common.CursorPaginatedResult[Conversation, string], error)
	GetPendingConversations(ctx context.Context, userID int64, q common.CursorQueryParams[string]) (common.CursorPaginatedResult[Conversation, string], error)

	CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*Conversation, error)
	CreateGroup(ctx context.Context, params CreateGroupParams) (*Conversation, error)
	GetConversation(ctx context.Context, convID string, userID int64) (*Conversation, error)
	UpdateGroup(ctx context.Context, params UpdateGroupParams) error
	AcceptConversation(ctx context.Context, convID string, userID int64) error
	IgnoreConversation(ctx context.Context, convID string, userID int64) error
	AddMember(ctx context.Context, convID string, actorID, newUserID int64) error
	RemoveMember(ctx context.Context, convID string, actorID, targetID int64) error
}

type MessageUseCase interface {
	GetMessages(ctx context.Context, params GetMessagesParams) (common.CursorPaginatedResult[Message, string], error)
	SendMessage(ctx context.Context, params SendMessageParams) (*Message, error)
	EditMessage(ctx context.Context, params EditMessageParams) (*Message, error)
	DeleteMessage(ctx context.Context, msgID string, userID int64) error
	AddReaction(ctx context.Context, msgID string, userID int64, emoji string) error
	RemoveReaction(ctx context.Context, msgID string, userID int64) error
	MarkRead(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type RealtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *Message, conv *Conversation) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

type UseCase struct {
	Conversations ConversationsUseCase
	Messages      MessageUseCase
}

type Deps struct {
	ConvRepo    ConversationRepository
	MsgRepo     MessageRepository
	Presence    PresenceStore
	Unread      UnreadStore
	Follow      FollowChecker
	Event       common.EventPublisher // RabbitMQ — publishes EventMessageCreated
	RTPublisher RealtimePublisher     // Redis PubSub — pushes real-time WS events
}
