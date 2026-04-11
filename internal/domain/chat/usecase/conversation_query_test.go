package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/chat"
	chatmocks "air-social/internal/domain/chat/mocks"
	"air-social/internal/domain/chat/usecase"
	"air-social/internal/domain/common"
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

func (s *conversationQuerySuite) TestGetConversations() {
	const userID = int64(1)

	// makeConvs tạo n conversation với UpdatedAt giảm dần (conv[0] mới nhất).
	makeConvs := func(n int) []chat.Conversation {
		convs := make([]chat.Conversation, n)
		base := time.Now().UTC().Truncate(time.Second)
		for i := range convs {
			convs[i] = chat.Conversation{
				ID:        fmt.Sprintf("01CONV%02d", i),
				UpdatedAt: base.Add(-time.Duration(i) * time.Minute),
			}
		}
		return convs
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

	// newUCNoUnread tạo usecase với Unread = nil — test nil-guard trong GetConversations.
	newUCNoUnread := func(d testDeps) *usecase.ConversationQueryUseCase {
		return usecase.NewQueryUseCase(usecase.QueryDeps{
			ConvRepo: d.convRepo,
			Unread:   nil,
		})
	}

	type testCase struct {
		name         string
		params       chat.GetConversationsParams
		setupMock    func(deps testDeps)
		buildUC      func(deps testDeps) *usecase.ConversationQueryUseCase
		wantErr      error
		assertResult func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string])
	}

	defaultParams := chat.GetConversationsParams{
		UserID: userID,
		// State empty → usecase default thành StateActive
		Query: common.CursorQueryParams[string]{Limit: 10},
	}

	tests := []testCase{
		{
			name: "repo_error",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			buildUC: newUC,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "unread_error",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(2), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(nil, assert.AnError).Once()
			},
			buildUC: newUC,
			wantErr: pkg.ErrInternal,
		},
		{
			// GetList trả 2 conv, limit=10 → không có trang tiếp.
			name:   "success_first_page_no_more",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(2), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			buildUC: newUC,
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 2)
				s.False(result.HasNextPage)
				s.Empty(result.NextCursor)
			},
		},
		{
			// GetList trả limit+1=11 conv → HasNextPage=true, data trimmed về 10.
			// NextCursor = UpdatedAt của item thứ 10 (index 9) vì GetCursor() = UpdatedAt.
			name: "success_has_next_page",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(11), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			buildUC: newUC,
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 10)
				s.True(result.HasNextPage)
				// cursor = GetCursor() của item cuối = UpdatedAt RFC3339Nano
				s.Equal(result.Data[9].GetCursor(), result.NextCursor)
			},
		},
		{
			// UnreadCount được populate đúng từ Redis map.
			name:   "success_populates_unread",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				convs := makeConvs(1)
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(convs, nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{convs[0].ID: 7}, nil).Once()
			},
			buildUC: newUC,
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Equal(7, result.Data[0].UnreadCount)
			},
		},
		{
			// Khi State="" → usecase tự default thành StateActive trước khi gọi repo.
			name:   "success_state_defaults_to_active",
			params: defaultParams, // State: ""
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.MatchedBy(func(p chat.GetConversationsParams) bool {
						return p.State == chat.StateActive
					})).
					Return(makeConvs(1), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			buildUC: newUC,
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 1)
			},
		},
		{
			// Khi Unread dep = nil → không panic, UnreadCount = 0 (zero value).
			name:   "success_nil_unread_store",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(2), nil).Once()
				// unreadStore không được gọi vì dep là nil
			},
			buildUC: newUCNoUnread,
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 2)
				for _, conv := range result.Data {
					s.Equal(0, conv.UnreadCount)
				}
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			deps := newDeps(s.T())
			uc := tc.buildUC(deps)

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetConversations(context.Background(), tc.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.Data)
			} else {
				s.NoError(err)
				if tc.assertResult != nil {
					tc.assertResult(s, got)
				}
			}
		})
	}
}
