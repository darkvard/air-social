package usecase

import (
	"context"
	"time"

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

	p := conv.FindParticipant(userID)
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

	p := conv.FindParticipant(followerID)
	if p == nil || p.State != chat.StatePending {
		return nil
	}

	return pkg.OrInternalError(u.deps.ConvRepo.UpdateParticipantState(ctx, conv.ID, followerID, chat.StateActive))
}

func (u *ConversationMemberUseCase) IgnoreConversation(ctx context.Context, convID string, userID int64) error {
	return nil
}

func (u *ConversationMemberUseCase) AddMember(ctx context.Context, p chat.AddMemberParams) error {
	conv, err := u.loadGroupConv(ctx, p.ConvID)
	if err != nil {
		return err
	}

	actor := conv.FindParticipant(p.ActorID)
	if actor == nil || !actor.IsAdmin() {
		return pkg.ErrForbidden
	}
	if conv.FindParticipant(p.NewUserID) != nil {
		return pkg.ErrAlreadyExists
	}

	rel, err := u.deps.FollowChecker.GetRelationship(ctx, p.ActorID, p.NewUserID)
	if err != nil {
		return pkg.OrInternalError(err)
	}

	state := chat.StatePending
	if rel.IsFollower {
		state = chat.StateActive
	}

	participant := chat.Participant{
		UserID:   p.NewUserID,
		Role:     chat.RoleMember,
		State:    state,
		JoinedAt: time.Now().UTC(),
	}
	return pkg.OrInternalError(u.deps.ConvRepo.AddParticipant(ctx, p.ConvID, participant))
}

func (u *ConversationMemberUseCase) RemoveMember(ctx context.Context, p chat.RemoveMemberParams) error {
	conv, err := u.loadGroupConv(ctx, p.ConvID)
	if err != nil {
		return err
	}

	actor := conv.FindParticipant(p.ActorID)
	if actor == nil {
		return pkg.ErrForbidden
	}

	if p.ActorID == p.TargetID {
		return pkg.OrInternalError(u.deps.ConvRepo.RemoveParticipant(ctx, p.ConvID, p.ActorID))
	}

	if !actor.IsAdmin() {
		return pkg.ErrForbidden
	}
	if conv.FindParticipant(p.TargetID) == nil {
		return pkg.ErrNotFound
	}
	return pkg.OrInternalError(u.deps.ConvRepo.RemoveParticipant(ctx, p.ConvID, p.TargetID))
}

func (u *ConversationMemberUseCase) UpdateMemberRole(ctx context.Context, p chat.UpdateMemberRoleParams) error {
	conv, err := u.loadGroupConv(ctx, p.ConvID)
	if err != nil {
		return err
	}

	actor := conv.FindParticipant(p.ActorID)
	if actor == nil || !actor.IsAdmin() {
		return pkg.ErrForbidden
	}
	if p.ActorID == p.TargetID {
		return pkg.NewError(pkg.ErrBadRequest, "cannot change own role")
	}
	if conv.FindParticipant(p.TargetID) == nil {
		return pkg.ErrNotFound
	}
	return pkg.OrInternalError(u.deps.ConvRepo.UpdateParticipantRole(ctx, p.ConvID, p.TargetID, p.Role))
}

// loadGroupConv fetches a conversation and verifies it exists and is a group.
func (u *ConversationMemberUseCase) loadGroupConv(ctx context.Context, convID string) (*chat.Conversation, error) {
	conv, err := u.deps.ConvRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil, pkg.ErrNotFound
	}
	if conv.Type == chat.ConversationDirect {
		return nil, pkg.NewError(pkg.ErrBadRequest, "group conversation required")
	}
	return conv, nil
}
