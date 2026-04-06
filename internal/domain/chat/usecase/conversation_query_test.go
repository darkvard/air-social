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
	"air-social/pkg"
)

type conversationQuerySuite struct {
	suite.Suite
}

func TestConversationQuerySuite(t *testing.T) {
	suite.Run(t, new(conversationQuerySuite))
}

func (s *conversationQuerySuite) TestGetConversation() {
	var (
		convID = "01CONV"
		userID = int64(1)
		other  = int64(2)
	)

	existingConv := func() *chat.Conversation {
		return &chat.Conversation{
			ID:   convID,
			Type: chat.ConversationDirect,
			Participants: []chat.Participant{
				{UserID: userID, Role: chat.RoleMember, State: chat.StateActive},
				{UserID: other, Role: chat.RoleMember, State: chat.StateActive},
			},
		}
	}

	type testDeps struct {
		convRepo    *chatmocks.MockConversationRepository
		unreadStore *chatmocks.MockUnreadStore
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:    chatmocks.NewMockConversationRepository(t),
			unreadStore: chatmocks.NewMockUnreadStore(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationQueryUseCase {
		return usecase.NewQueryUseCase(usecase.QueryDeps{
			ConvRepo: d.convRepo,
			Unread:   d.unreadStore,
		})
	}

	tests := []struct {
		name         string
		setupMock    func(deps testDeps)
		wantErr      error
		assertResult func(s *conversationQuerySuite, conv *chat.Conversation)
	}{
		{
			name: "repo_error",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "not_found",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "not_member",
			setupMock: func(deps testDeps) {
				conv := existingConv()
				conv.Participants = []chat.Participant{
					{UserID: other, Role: chat.RoleMember, State: chat.StateActive},
				}
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(conv, nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "unread_error",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(existingConv(), nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(0), assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_with_unread_count",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(existingConv(), nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(5), nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, conv *chat.Conversation) {
				s.Equal(convID, conv.ID)
				s.Equal(5, conv.UnreadCount)
				s.Len(conv.Participants, 2)
			},
		},
		{
			name: "success_unread_zero",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(existingConv(), nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(0), nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, conv *chat.Conversation) {
				s.Equal(0, conv.UnreadCount)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := newDeps(s.T())
			uc := newUC(deps)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetConversation(context.Background(), convID, userID)

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
