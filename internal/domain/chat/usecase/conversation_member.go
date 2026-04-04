package usecase

import (
	"context"

	"air-social/internal/domain/chat"
)

type MemberDeps struct {
	ConvRepo      chat.ConversationRepository
	FollowChecker FollowChecker
}

type ConversationMemberUseCase struct {
	deps MemberDeps
}

func NewMemberUseCase(d MemberDeps) *ConversationMemberUseCase {
	return &ConversationMemberUseCase{deps: d}
}

func (u *ConversationMemberUseCase) AcceptConversation(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (u *ConversationMemberUseCase) IgnoreConversation(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (u *ConversationMemberUseCase) AddMember(ctx context.Context, convID string, actorID, newUserID int64) error {
	return nil
}

func (u *ConversationMemberUseCase) RemoveMember(ctx context.Context, convID string, actorID, targetID int64) error {
	return nil
}
