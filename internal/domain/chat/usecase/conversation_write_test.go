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

type conversationWriteSuite struct {
	suite.Suite
}

func TestConversationWriteSuite(t *testing.T) {
	suite.Run(t, new(conversationWriteSuite))
}

func (s *conversationWriteSuite) TestCreateOrGetDirect() {
	var (
		senderID    = int64(1)
		recipientID = int64(2)
		lastMsgID   = "01MSG"
	)

	type testDeps struct {
		convRepo      *chatmocks.MockConversationRepository
		msgRepo       *chatmocks.MockMessageRepository
		unread        *chatmocks.MockUnreadStore
		followChecker *ucmocks.MockFollowChecker
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:      chatmocks.NewMockConversationRepository(t),
			msgRepo:       chatmocks.NewMockMessageRepository(t),
			unread:        chatmocks.NewMockUnreadStore(t),
			followChecker: ucmocks.NewMockFollowChecker(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationWriteUseCase {
		return usecase.NewWriteUseCase(usecase.WriteDeps{
			ConvRepo:      d.convRepo,
			MsgRepo:       d.msgRepo,
			Unread:        d.unread,
			FollowChecker: d.followChecker,
		})
	}

	tests := []struct {
		name         string
		setupMock    func(deps testDeps)
		wantErr      error
		wantNew      bool
		assertResult func(s *conversationWriteSuite, conv *chat.Conversation)
	}{
		{
			name: "find_direct_error",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "existing_conv_no_last_msg",
			setupMock: func(deps testDeps) {
				existing := &chat.Conversation{ID: "01EXISTING", Type: chat.ConversationDirect}
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(existing, nil).Once()
				deps.unread.EXPECT().
					Get(mock.Anything, senderID, "01EXISTING").
					Return(int64(3), nil).Once()
			},
			wantNew: false,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("01EXISTING", conv.ID)
				s.Equal(3, conv.UnreadCount)
				s.Nil(conv.LastMessage)
			},
		},
		{
			name: "existing_conv_with_last_msg",
			setupMock: func(deps testDeps) {
				existing := &chat.Conversation{ID: "01EXISTING", Type: chat.ConversationDirect, LastMsgID: lastMsgID}
				lastMsg := chat.Message{ID: lastMsgID, Content: "hello"}
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(existing, nil).Once()
				deps.msgRepo.EXPECT().
					GetByIDs(mock.Anything, []string{lastMsgID}).
					Return([]chat.Message{lastMsg}, nil).Once()
				deps.unread.EXPECT().
					Get(mock.Anything, senderID, "01EXISTING").
					Return(int64(1), nil).Once()
			},
			wantNew: false,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("01EXISTING", conv.ID)
				s.Require().NotNil(conv.LastMessage)
				s.Equal(lastMsgID, conv.LastMessage.ID)
				s.Equal(1, conv.UnreadCount)
			},
		},
		{
			name: "follow_checker_error",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "creates_with_pending_state",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{IsFollowing: false, IsFollower: false}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal(chat.ConversationDirect, conv.Type)
				s.Equal(senderID, conv.CreatedBy)
				s.Len(conv.Participants, 2)
				senderP, recipientP := conv.Participants[0], conv.Participants[1]
				s.Equal(senderID, senderP.UserID)
				s.Equal(chat.StateActive, senderP.State)
				s.Equal(recipientID, recipientP.UserID)
				s.Equal(chat.StatePending, recipientP.State)
			},
		},
		{
			name: "creates_with_active_state",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{IsFollowing: false, IsFollower: true}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal(chat.ConversationDirect, conv.Type)
				recipientP := conv.Participants[1]
				s.Equal(recipientID, recipientP.UserID)
				s.Equal(chat.StateActive, recipientP.State)
			},
		},
		{
			name: "create_repo_error",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := newDeps(s.T())
			uc := newUC(deps)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, isNew, err := uc.CreateOrGetDirect(context.Background(), senderID, recipientID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.NotNil(got)
				s.Equal(tc.wantNew, isNew)
				if tc.assertResult != nil {
					tc.assertResult(s, got)
				}
			}
		})
	}
}

func (s *conversationWriteSuite) TestCreateGroup() {
	var (
		creatorID = int64(1)
		member1   = int64(2)
		member2   = int64(3)
		member3   = int64(4)
	)

	type testDeps struct {
		convRepo      *chatmocks.MockConversationRepository
		msgRepo       *chatmocks.MockMessageRepository
		unread        *chatmocks.MockUnreadStore
		followChecker *ucmocks.MockFollowChecker
		userChecker   *ucmocks.MockUserChecker
		mediaVerifier *ucmocks.MockMediaVerifier
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:      chatmocks.NewMockConversationRepository(t),
			msgRepo:       chatmocks.NewMockMessageRepository(t),
			unread:        chatmocks.NewMockUnreadStore(t),
			followChecker: ucmocks.NewMockFollowChecker(t),
			userChecker:   ucmocks.NewMockUserChecker(t),
			mediaVerifier: ucmocks.NewMockMediaVerifier(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationWriteUseCase {
		return usecase.NewWriteUseCase(usecase.WriteDeps{
			ConvRepo:      d.convRepo,
			MsgRepo:       d.msgRepo,
			Unread:        d.unread,
			FollowChecker: d.followChecker,
			UserChecker:   d.userChecker,
			MediaVerifier: d.mediaVerifier,
		})
	}

	noFollows := func(deps testDeps, memberIDs []int64) {
		rels := make(map[int64]follow.Relationship, len(memberIDs))
		for _, id := range memberIDs {
			rels[id] = follow.Relationship{IsFollower: false}
		}
		deps.followChecker.EXPECT().
			GetRelationships(mock.Anything, creatorID, memberIDs).
			Return(rels, nil).Once()
	}

	tests := []struct {
		name         string
		params       chat.CreateGroupParams
		setupMock    func(deps testDeps)
		wantErr      error
		wantNew      bool
		assertResult func(s *conversationWriteSuite, conv *chat.Conversation)
	}{
		{
			name: "too_few_members_after_dedup",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member1},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "creator_excluded_from_members",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{creatorID, member1},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {},
			wantErr:   pkg.ErrBadRequest,
		},
		{
			name: "avatar_key_verify_error",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
				AvatarKey: "groups/tmp/avatar/key.jpg",
			},
			setupMock: func(deps testDeps) {
				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"groups/tmp/avatar/key.jpg"}).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "avatar_key_not_found",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
				AvatarKey: "groups/tmp/avatar/key.jpg",
			},
			setupMock: func(deps testDeps) {
				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"groups/tmp/avatar/key.jpg"}).
					Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "idempotency_existing_conv_returned",
			params: chat.CreateGroupParams{
				CreatorID:    creatorID,
				MemberIDs:    []int64{member1, member2},
				Name:         "Team",
				ClientConvID: "client-uuid-123",
			},
			setupMock: func(deps testDeps) {
				existing := &chat.Conversation{ID: "01EXISTING", Type: chat.ConversationGroup}
				deps.convRepo.EXPECT().
					GetByClientConvID(mock.Anything, "client-uuid-123").
					Return(existing, nil).Once()
				// LastMsgID="" → skip MsgRepo; Unread.Get always called
				deps.unread.EXPECT().
					Get(mock.Anything, creatorID, "01EXISTING").
					Return(int64(0), nil).Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("01EXISTING", conv.ID)
			},
		},
		{
			name: "user_checker_error",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "member_ids_not_found",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return([]int64{member2}, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "member_active_if_follows_creator",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationships(mock.Anything, creatorID, []int64{member1, member2}).
					Return(map[int64]follow.Relationship{
						member1: {IsFollower: true},
						member2: {IsFollower: true},
					}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				pm := participantMap(conv)
				s.Equal(chat.StateActive, pm[member1].State)
				s.Equal(chat.StateActive, pm[member2].State)
			},
		},
		{
			name: "member_pending_if_not_following_creator",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationships(mock.Anything, creatorID, []int64{member1, member2}).
					Return(map[int64]follow.Relationship{
						member1: {IsFollower: false},
						member2: {IsFollower: false},
					}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				pm := participantMap(conv)
				s.Equal(chat.StatePending, pm[member1].State)
				s.Equal(chat.StatePending, pm[member2].State)
			},
		},
		{
			name: "mixed_states_per_member",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationships(mock.Anything, creatorID, []int64{member1, member2}).
					Return(map[int64]follow.Relationship{
						member1: {IsFollower: true},
						member2: {IsFollower: false},
					}, nil).Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				pm := participantMap(conv)
				s.Equal(chat.StateActive, pm[member1].State)
				s.Equal(chat.StatePending, pm[member2].State)
			},
		},
		{
			name: "follow_checker_error",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				deps.followChecker.EXPECT().
					GetRelationships(mock.Anything, creatorID, []int64{member1, member2}).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "creates_successfully",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				noFollows(deps, []int64{member1, member2})
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal(chat.ConversationGroup, conv.Type)
				s.Equal("Team", conv.Name)
				s.Equal(creatorID, conv.CreatedBy)
				s.Len(conv.Participants, 3)
				creator := conv.Participants[0]
				s.Equal(creatorID, creator.UserID)
				s.Equal(chat.RoleAdmin, creator.Role)
				s.Equal(chat.StateActive, creator.State)
				s.Equal(chat.RoleMember, conv.Participants[1].Role)
				s.Equal(chat.RoleMember, conv.Participants[2].Role)
			},
		},
		{
			name: "dedup_members_and_creates",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2, member1, member3},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2, member3}).
					Return(nil, nil).Once()
				noFollows(deps, []int64{member1, member2, member3})
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Len(conv.Participants, 4)
			},
		},
		{
			name: "with_avatar_key_creates_successfully",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
				AvatarKey: "groups/tmp/avatar/key.jpg",
			},
			setupMock: func(deps testDeps) {
				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"groups/tmp/avatar/key.jpg"}).
					Return(nil).Once()
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				noFollows(deps, []int64{member1, member2})
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).Once()
			},
			wantNew: true,
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("groups/tmp/avatar/key.jpg", conv.AvatarKey)
			},
		},
		{
			name: "idempotency_conflict_on_create_fetches_existing",
			params: chat.CreateGroupParams{
				CreatorID:    creatorID,
				MemberIDs:    []int64{member1, member2},
				Name:         "Team",
				ClientConvID: "client-uuid-456",
			},
			setupMock: func(deps testDeps) {
				existing := &chat.Conversation{ID: "01RACE", Type: chat.ConversationGroup}
				deps.convRepo.EXPECT().
					GetByClientConvID(mock.Anything, "client-uuid-456").
					Return(nil, nil).Once()
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				noFollows(deps, []int64{member1, member2})
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(pkg.ErrConflict).Once()
				deps.convRepo.EXPECT().
					GetByClientConvID(mock.Anything, "client-uuid-456").
					Return(existing, nil).Once()
				deps.unread.EXPECT().
					Get(mock.Anything, creatorID, "01RACE").
					Return(int64(0), nil).Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("01RACE", conv.ID)
			},
		},
		{
			name: "create_repo_error",
			params: chat.CreateGroupParams{
				CreatorID: creatorID,
				MemberIDs: []int64{member1, member2},
				Name:      "Team",
			},
			setupMock: func(deps testDeps) {
				deps.userChecker.EXPECT().
					FilterNonExistent(mock.Anything, []int64{member1, member2}).
					Return(nil, nil).Once()
				noFollows(deps, []int64{member1, member2})
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := newDeps(s.T())
			uc := newUC(deps)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, isNew, err := uc.CreateGroup(context.Background(), tc.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.NotNil(got)
				s.Equal(tc.wantNew, isNew)
				if tc.assertResult != nil {
					tc.assertResult(s, got)
				}
			}
		})
	}
}

func (s *conversationWriteSuite) TestUpdateGroup() {
	var (
		convID  = "01CONV"
		actorID = int64(1)
	)

	strPtr := func(v string) *string { return &v }

	type testDeps struct {
		convRepo      *chatmocks.MockConversationRepository
		mediaVerifier *ucmocks.MockMediaVerifier
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:      chatmocks.NewMockConversationRepository(t),
			mediaVerifier: ucmocks.NewMockMediaVerifier(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationWriteUseCase {
		return usecase.NewWriteUseCase(usecase.WriteDeps{
			ConvRepo:      d.convRepo,
			MediaVerifier: d.mediaVerifier,
		})
	}

	tests := []struct {
		name         string
		params       chat.UpdateGroupParams
		setupMock    func(d testDeps)
		wantErr      error
		assertResult func(s *conversationWriteSuite, conv *chat.Conversation)
	}{
		{
			name:   "repo_get_error",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("X")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "conv_not_found",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("X")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "direct_conv_rejected",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("X")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).Return(directConv(convID), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:   "actor_not_in_conv",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("X")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(99)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "actor_not_admin",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("X")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, member(actorID)), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "nothing_to_update",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:   "avatar_verify_error",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, AvatarKey: strPtr("key.jpg")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"key.jpg"}).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "avatar_not_found",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, AvatarKey: strPtr("key.jpg")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"key.jpg"}).
					Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:   "repo_update_error",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("New")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.convRepo.EXPECT().UpdateGroupInfo(mock.Anything, convID, strPtr("New"), (*string)(nil)).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "updates_name_only",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, Name: strPtr("New Name")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.convRepo.EXPECT().UpdateGroupInfo(mock.Anything, convID, strPtr("New Name"), (*string)(nil)).
					Return(nil).Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("New Name", conv.Name)
			},
		},
		{
			name:   "updates_avatar_only",
			params: chat.UpdateGroupParams{ConvID: convID, ActorID: actorID, AvatarKey: strPtr("new/avatar.jpg")},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"new/avatar.jpg"}).Return(nil).Once()
				d.convRepo.EXPECT().UpdateGroupInfo(mock.Anything, convID, (*string)(nil), strPtr("new/avatar.jpg")).
					Return(nil).Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("new/avatar.jpg", conv.AvatarKey)
			},
		},
		{
			name: "updates_both_fields",
			params: chat.UpdateGroupParams{
				ConvID: convID, ActorID: actorID,
				Name: strPtr("Team"), AvatarKey: strPtr("team/avatar.jpg"),
			},
			setupMock: func(d testDeps) {
				d.convRepo.EXPECT().GetByID(mock.Anything, convID).
					Return(groupConv(convID, admin(actorID)), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"team/avatar.jpg"}).Return(nil).Once()
				d.convRepo.EXPECT().UpdateGroupInfo(mock.Anything, convID, strPtr("Team"), strPtr("team/avatar.jpg")).
					Return(nil).Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal("Team", conv.Name)
				s.Equal("team/avatar.jpg", conv.AvatarKey)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			d := newDeps(s.T())
			tc.setupMock(d)
			got, err := newUC(d).UpdateGroup(context.Background(), tc.params)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.NotNil(got)
				if tc.assertResult != nil {
					tc.assertResult(s, got)
				}
			}
		})
	}
}

func participantMap(conv *chat.Conversation) map[int64]chat.Participant {
	m := make(map[int64]chat.Participant, len(conv.Participants))
	for _, p := range conv.Participants {
		m[p.UserID] = p
	}
	return m
}
