package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
	convo "air-social/internal/domain/chat/usecase/conversation"
	"air-social/internal/domain/chat/usecase/messages"
)

// ConversationQuerier, ConversationWriter, ConversationMember, Messenger —
// declared here because ConversationUseCase/UseCase structs are the consumers.

type ConversationQuerier interface {
	GetConversations(ctx context.Context, params chat.GetConversationsParams) (common.CursorPaginatedResult[chat.Conversation, string], error)
	GetPendingConversations(ctx context.Context, userID int64, q common.CursorQueryParams[string]) (common.CursorPaginatedResult[chat.Conversation, string], error)
	GetConversation(ctx context.Context, convID string, userID int64) (*chat.Conversation, error)
}

type ConversationWriter interface {
	CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*chat.Conversation, error)
	CreateGroup(ctx context.Context, params chat.CreateGroupParams) (*chat.Conversation, error)
	UpdateGroup(ctx context.Context, params chat.UpdateGroupParams) error
}

type ConversationMember interface {
	AcceptConversation(ctx context.Context, convID string, userID int64) error
	IgnoreConversation(ctx context.Context, convID string, userID int64) error
	AddMember(ctx context.Context, convID string, actorID, newUserID int64) error
	RemoveMember(ctx context.Context, convID string, actorID, targetID int64) error
}

type Messenger interface {
	GetMessages(ctx context.Context, params chat.GetMessagesParams) (common.CursorPaginatedResult[chat.Message, string], error)
	SendMessage(ctx context.Context, params chat.SendMessageParams) (*chat.Message, error)
	EditMessage(ctx context.Context, params chat.EditMessageParams) (*chat.Message, error)
	DeleteMessage(ctx context.Context, msgID string, userID int64) error
	AddReaction(ctx context.Context, msgID string, userID int64, emoji string) error
	RemoveReaction(ctx context.Context, msgID string, userID int64) error
	MarkRead(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

// FollowChecker and RealtimePublisher — declared here because Deps/New() consume them.

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type RealtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
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

type Deps struct {
	ConvRepo    chat.ConversationRepository
	MsgRepo     chat.MessageRepository
	Presence    chat.PresenceStore
	Unread      chat.UnreadStore
	Follow      FollowChecker
	Event       common.EventPublisher
	RTPublisher RealtimePublisher
}

func New(d Deps) UseCase {
	return UseCase{
		Conversation: ConversationUseCase{
			Query:  convo.NewQueryUseCase(d.ConvRepo, d.Unread),
			Write:  convo.NewWriteUseCase(d.ConvRepo, d.Follow, d.Event, d.RTPublisher),
			Member: convo.NewMemberUseCase(d.ConvRepo, d.Follow),
		},
		Messages: messages.NewMessageUseCase(d.MsgRepo, d.ConvRepo, d.Event, d.RTPublisher, d.Unread),
	}
}
