package conversation

import (
	"context"

	"air-social/internal/domain/chat"
)

// followChecker is declared in write.go (same package) and reused here.

type MemberUseCase struct {
	convRepo      chat.ConversationRepository
	followChecker followChecker
}

func NewMemberUseCase(convRepo chat.ConversationRepository, fc followChecker) *MemberUseCase {
	return &MemberUseCase{convRepo: convRepo, followChecker: fc}
}

func (u *MemberUseCase) AcceptConversation(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (u *MemberUseCase) IgnoreConversation(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (u *MemberUseCase) AddMember(ctx context.Context, convID string, actorID, newUserID int64) error {
	return nil
}

func (u *MemberUseCase) RemoveMember(ctx context.Context, convID string, actorID, targetID int64) error {
	return nil
}
