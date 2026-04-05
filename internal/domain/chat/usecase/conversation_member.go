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

// AcceptConversation moves the caller's participant state from pending/ignored → active.
// pending → active has 2 triggers:
//  1. Automatic: EventFollowCreated worker detects caller followed the sender and updates state.
//  2. Manual:    Caller explicitly accepts via this method (e.g. user taps "Accept" in UI).
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
