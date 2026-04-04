package conversation

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
)

// Declared here because WriteUseCase is the consumer of these dependencies.

type followChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type realtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

type WriteUseCase struct {
	convRepo      chat.ConversationRepository
	followChecker followChecker
	event         common.EventPublisher
	rtPublisher   realtimePublisher
}

func NewWriteUseCase(
	convRepo chat.ConversationRepository,
	followChecker followChecker,
	event common.EventPublisher,
	rtPublisher realtimePublisher,
) *WriteUseCase {
	return &WriteUseCase{
		convRepo:      convRepo,
		followChecker: followChecker,
		event:         event,
		rtPublisher:   rtPublisher,
	}
}

func (u *WriteUseCase) CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*chat.Conversation, error) {
	return nil, nil
}

func (u *WriteUseCase) CreateGroup(ctx context.Context, params chat.CreateGroupParams) (*chat.Conversation, error) {
	return nil, nil
}

func (u *WriteUseCase) UpdateGroup(ctx context.Context, params chat.UpdateGroupParams) error {
	return nil
}
