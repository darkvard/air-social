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
	"air-social/internal/domain/follow"
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
	)

	existingConv := &chat.Conversation{
		ID:   "01EXISTING",
		Type: chat.ConversationDirect,
	}

	type testDeps struct {
		convRepo      *chatmocks.MockConversationRepository
		followChecker *chatmocks.MockFollowChecker
	}

	type args struct {
		ctx         context.Context
		senderID    int64
		recipientID int64
	}

	tests := []struct {
		name         string
		args         args
		setupMock    func(deps testDeps)
		wantErr      error
		assertResult func(s *conversationWriteSuite, conv *chat.Conversation)
	}{
		{
			name: "find_direct_error",
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "existing_conv_returned",
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(existingConv, nil).
					Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal(existingConv.ID, conv.ID)
			},
		},
		{
			name: "follow_checker_error",
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).
					Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{}, assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "creates_with_pending_state",
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).
					Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{IsFollowing: false, IsFollower: false}, nil).
					Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).
					Once()
			},
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
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).
					Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{IsFollowing: false, IsFollower: true}, nil).
					Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(nil).
					Once()
			},
			assertResult: func(s *conversationWriteSuite, conv *chat.Conversation) {
				s.Equal(chat.ConversationDirect, conv.Type)
				recipientP := conv.Participants[1]
				s.Equal(recipientID, recipientP.UserID)
				s.Equal(chat.StateActive, recipientP.State)
			},
		},
		{
			name: "create_repo_error",
			args: args{ctx: context.Background(), senderID: senderID, recipientID: recipientID},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					FindDirect(mock.Anything, senderID, recipientID).
					Return(nil, nil).
					Once()
				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, senderID, recipientID).
					Return(follow.Relationship{}, nil).
					Once()
				deps.convRepo.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*chat.Conversation")).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockConvRepo := chatmocks.NewMockConversationRepository(s.T())
			mockFollowChecker := chatmocks.NewMockFollowChecker(s.T())

			deps := testDeps{
				convRepo:      mockConvRepo,
				followChecker: mockFollowChecker,
			}

			uc := usecase.NewWriteUseCase(usecase.WriteDeps{
				ConvRepo:      mockConvRepo,
				FollowChecker: mockFollowChecker,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.CreateOrGetDirect(tc.args.ctx, tc.args.senderID, tc.args.recipientID)

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
