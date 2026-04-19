package usecase

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/pkg"
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

// AcceptConversation moves the caller's participant state from pending/ignored -> active.
func (u *ConversationMemberUseCase) AcceptConversation(ctx context.Context, convID string, userID int64) error {
	conv, err := u.deps.ConvRepo.GetByID(ctx, convID)
	if err != nil {
		return pkg.OrInternalError(err)
	}
	if conv == nil {
		return pkg.ErrNotFound
	}

	p := findParticipant(conv.Participants, userID)
	if p == nil {
		return pkg.ErrForbidden
	}
	if p.State == chat.StateActive {
		return pkg.NewError(pkg.ErrBadRequest, "already active")
	}

	return pkg.OrInternalError(u.deps.ConvRepo.UpdateParticipantState(ctx, convID, userID, chat.StateActive))
}

// AcceptPendingByFollowEvent is called by the follow-event worker.
// When B follows A, if B has a pending direct conv with A, auto-accept it silently.
func (u *ConversationMemberUseCase) AcceptPendingByFollowEvent(ctx context.Context, followerID, followeeID int64) error {
	conv, err := u.deps.ConvRepo.GetDirect(ctx, followerID, followeeID)
	if err != nil {
		return pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil
	}

	p := findParticipant(conv.Participants, followerID)
	if p == nil || p.State != chat.StatePending {
		return nil
	}

	return pkg.OrInternalError(u.deps.ConvRepo.UpdateParticipantState(ctx, conv.ID, followerID, chat.StateActive))
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

func (u *ConversationMemberUseCase) UpdateMemberRole(ctx context.Context, convID string, actorID, targetID int64, role chat.ParticipantRole) error {
	return nil
}
