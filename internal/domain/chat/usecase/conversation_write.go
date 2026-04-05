package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
)

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type RealtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
	PublishMessageDeleted(ctx context.Context, msgID string, convID string) error
	PublishReactionAdded(ctx context.Context, msgID string, convID string, reaction chat.Reaction) error
	PublishReactionRemoved(ctx context.Context, msgID string, convID string, userID int64) error
}

type WriteDeps struct {
	ConvRepo      chat.ConversationRepository
	FollowChecker FollowChecker
	Event         common.EventPublisher
	RTPublisher   RealtimePublisher
}

type ConversationWriteUseCase struct {
	deps WriteDeps
}

func NewWriteUseCase(d WriteDeps) *ConversationWriteUseCase {
	return &ConversationWriteUseCase{deps: d}
}

func (u *ConversationWriteUseCase) CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*chat.Conversation, error) {
	existingConv, err := u.deps.ConvRepo.FindDirect(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}
	if existingConv != nil {
		// TODO: populate existingConv.LastMessage from MessageRepository once message layer is implemented.
		return existingConv, nil
	}

	relationship, err := u.deps.FollowChecker.GetRelationship(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}

	newConv := chat.NewDirectConversation(senderID, recipientID, relationship.IsFollower)
	if err := u.deps.ConvRepo.Create(ctx, newConv); err != nil {
		return nil, err
	}
	return newConv, nil
}

func (u *ConversationWriteUseCase) CreateGroup(ctx context.Context, params chat.CreateGroupParams) (*chat.Conversation, error) {
	return nil, nil
}

func (u *ConversationWriteUseCase) UpdateGroup(ctx context.Context, params chat.UpdateGroupParams) error {
	return nil
}