package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/chat"
	chatmocks "air-social/internal/domain/chat/mocks"
	"air-social/internal/domain/chat/usecase"
	ucmocks "air-social/internal/domain/chat/usecase/mocks"
	"air-social/internal/domain/follow"
	"air-social/pkg"
)

type conversationMemberSuite struct {
	suite.Suite
}

func TestConversationMemberSuite(t *testing.T) {
	suite.Run(t, new(conversationMemberSuite))
}

// ── helpers ──────────────────────────────────────────────────────────────────

func groupConv(convID string, participants ...chat.Participant) *chat.Conversation {
	return &chat.Conversation{
		ID:           convID,
		Type:         chat.ConversationGroup,
		Participants: participants,
	}
}

func directConv(convID string) *chat.Conversation {
	return &chat.Conversation{ID: convID, Type: chat.ConversationDirect}
}

func admin(userID int64) chat.Participant {
	return chat.Participant{UserID: userID, Role: chat.RoleAdmin, State: chat.StateActive}
}

func member(userID int64) chat.Participant {
	return chat.Participant{UserID: userID, Role: chat.RoleMember, State: chat.StateActive}
}

func pendingMember(userID int64) chat.Participant {
	return chat.Participant{UserID: userID, Role: chat.RoleMember, State: chat.StatePending}
}

func ignoredMember(userID int64) chat.Participant {
	return chat.Participant{UserID: userID, Role: chat.RoleMember, State: chat.StateIgnored}
}

type memberDeps struct {
	repo          *chatmocks.MockConversationRepository
	followChecker *ucmocks.MockFollowChecker
}

func newDeps(t *testing.T) memberDeps {
	return memberDeps{
		repo:          chatmocks.NewMockConversationRepository(t),
		followChecker: ucmocks.NewMockFollowChecker(t),
	}
}

func newMemberUC(d memberDeps) *usecase.ConversationMemberUseCase {
	return usecase.NewMemberUseCase(usecase.MemberDeps{
		ConvRepo:      d.repo,
		FollowChecker: d.followChecker,
	})
}

// ── AcceptConversation ────────────────────────────────────────────────────────

func (s *conversationMemberSuite) TestAcceptConversation() {
	var (
		convID = "01CONV"
		userID = int64(1)
	)

	tests := []struct {
		name      string
		setupMock func(d memberDeps)
		wantErr   error
	}{
		{
			name: "repo_get_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "conv_not_found",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "user_not_participant",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(99)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "already_active",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(userID)), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "accept_from_pending",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, pendingMember(userID)), nil).Once()
				d.repo.EXPECT().UpdateParticipantState(mock.Anything, convID, userID, chat.StateActive).Return(nil).Once()
			},
		},
		{
			name: "accept_from_ignored",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, ignoredMember(userID)), nil).Once()
				d.repo.EXPECT().UpdateParticipantState(mock.Anything, convID, userID, chat.StateActive).Return(nil).Once()
			},
		},
		{
			name: "update_state_repo_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, pendingMember(userID)), nil).Once()
				d.repo.EXPECT().UpdateParticipantState(mock.Anything, convID, userID, chat.StateActive).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := newDeps(s.T())
			tc.setupMock(d)
			err := newMemberUC(d).AcceptConversation(context.Background(), convID, userID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── AcceptPendingByFollowEvent ────────────────────────────────────────────────

func (s *conversationMemberSuite) TestAcceptPendingByFollowEvent() {
	var (
		followerID = int64(2) // B just followed A
		followeeID = int64(1) // A
		convID     = "01CONV"
	)

	tests := []struct {
		name      string
		setupMock func(d memberDeps)
		wantErr   error
	}{
		{
			name: "get_direct_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "no_conv_silent_skip",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(nil, nil).Once()
			},
		},
		{
			name: "follower_not_participant_silent_skip",
			setupMock: func(d memberDeps) {
				conv := &chat.Conversation{ID: convID, Participants: []chat.Participant{member(followeeID)}}
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(conv, nil).Once()
			},
		},
		{
			name: "already_active_silent_skip",
			setupMock: func(d memberDeps) {
				conv := &chat.Conversation{ID: convID, Participants: []chat.Participant{member(followerID)}}
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(conv, nil).Once()
			},
		},
		{
			name: "pending_accepted",
			setupMock: func(d memberDeps) {
				conv := &chat.Conversation{ID: convID, Participants: []chat.Participant{pendingMember(followerID)}}
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(conv, nil).Once()
				d.repo.EXPECT().UpdateParticipantState(mock.Anything, convID, followerID, chat.StateActive).Return(nil).Once()
			},
		},
		{
			name: "update_state_error",
			setupMock: func(d memberDeps) {
				conv := &chat.Conversation{ID: convID, Participants: []chat.Participant{pendingMember(followerID)}}
				d.repo.EXPECT().GetDirect(mock.Anything, followerID, followeeID).Return(conv, nil).Once()
				d.repo.EXPECT().UpdateParticipantState(mock.Anything, convID, followerID, chat.StateActive).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := newDeps(s.T())
			tc.setupMock(d)
			err := newMemberUC(d).AcceptPendingByFollowEvent(context.Background(), followerID, followeeID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── AddMember ─────────────────────────────────────────────────────────────────

func (s *conversationMemberSuite) TestAddMember() {
	var (
		convID  = "01CONV"
		actorID = int64(1)
		newUser = int64(3)
	)

	tests := []struct {
		name      string
		setupMock func(d memberDeps)
		wantErr   error
	}{
		{
			name: "repo_get_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "conv_not_found",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "direct_conv_rejected",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(directConv(convID), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "actor_not_in_conv",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(99)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "actor_not_admin",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(actorID), member(newUser-1)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "user_already_member",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID), member(newUser)), nil).Once()
			},
			wantErr: pkg.ErrAlreadyExists,
		},
		{
			name: "follow_checker_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.followChecker.EXPECT().
					GetRelationship(mock.Anything, actorID, newUser).
					Return(follow.Relationship{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "adds_with_active_if_follows_actor",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.followChecker.EXPECT().
					GetRelationship(mock.Anything, actorID, newUser).
					Return(follow.Relationship{IsFollower: true}, nil).Once()
				d.repo.EXPECT().AddParticipant(mock.Anything, convID, mock.MatchedBy(func(p chat.Participant) bool {
					return p.UserID == newUser && p.State == chat.StateActive
				})).Return(nil).Once()
			},
		},
		{
			name: "adds_with_pending_if_not_following",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.followChecker.EXPECT().
					GetRelationship(mock.Anything, actorID, newUser).
					Return(follow.Relationship{IsFollower: false}, nil).Once()
				d.repo.EXPECT().AddParticipant(mock.Anything, convID, mock.MatchedBy(func(p chat.Participant) bool {
					return p.UserID == newUser && p.State == chat.StatePending
				})).Return(nil).Once()
			},
		},
		{
			name: "add_participant_repo_error",
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.followChecker.EXPECT().
					GetRelationship(mock.Anything, actorID, newUser).
					Return(follow.Relationship{}, nil).Once()
				d.repo.EXPECT().AddParticipant(mock.Anything, convID, mock.AnythingOfType("chat.Participant")).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := newDeps(s.T())
			tc.setupMock(d)
			err := newMemberUC(d).AddMember(context.Background(), convID, actorID, newUser)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── RemoveMember ──────────────────────────────────────────────────────────────

func (s *conversationMemberSuite) TestRemoveMember() {
	var (
		convID   = "01CONV"
		actorID  = int64(1)
		targetID = int64(2)
	)

	tests := []struct {
		name      string
		actorID   int64
		targetID  int64
		setupMock func(d memberDeps)
		wantErr   error
	}{
		{
			name:     "repo_get_error",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:     "conv_not_found",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:     "direct_conv_rejected",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).Return(directConv(convID), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:     "actor_not_in_conv",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(99)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:     "self_leave_success",
			actorID:  actorID,
			targetID: actorID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(actorID), member(targetID)), nil).Once()
				d.repo.EXPECT().RemoveParticipant(mock.Anything, convID, actorID).Return(nil).Once()
			},
		},
		{
			name:     "non_admin_cannot_remove_other",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(actorID), member(targetID)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:     "target_not_found",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:     "admin_removes_member_success",
			actorID:  actorID,
			targetID: targetID,
			setupMock: func(d memberDeps) {
				d.repo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID), member(targetID)), nil).Once()
				d.repo.EXPECT().RemoveParticipant(mock.Anything, convID, targetID).Return(nil).Once()
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := newDeps(s.T())
			tc.setupMock(d)
			err := newMemberUC(d).RemoveMember(context.Background(), convID, tc.actorID, tc.targetID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

