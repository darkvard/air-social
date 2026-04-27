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
		msgRepo     *chatmocks.MockMessageRepository
		unreadStore *chatmocks.MockUnreadStore
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:    chatmocks.NewMockConversationRepository(t),
			msgRepo:     chatmocks.NewMockMessageRepository(t),
			unreadStore: chatmocks.NewMockUnreadStore(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationQueryUseCase {
		return usecase.NewQueryUseCase(usecase.QueryDeps{
			ConvRepo: d.convRepo,
			MsgRepo:  d.msgRepo,
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
					Return(nil, nil).Once()
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
			wantErr: pkg.ErrForbidden,
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
				s.Nil(conv.LastMessage)
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
		{
			// conv has LastMsgID → MsgRepo.GetByIDs called → LastMessage populated.
			name: "success_populates_last_message",
			setupMock: func(deps testDeps) {
				conv := existingConv()
				conv.LastMsgID = "01MSG"
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(conv, nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(2), nil).Once()
				deps.msgRepo.EXPECT().
					GetByIDs(mock.Anything, []string{"01MSG"}).
					Return([]chat.Message{{ID: "01MSG", Content: "hello"}}, nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, conv *chat.Conversation) {
				s.NotNil(conv.LastMessage)
				s.Equal("01MSG", conv.LastMessage.ID)
				s.Equal("hello", conv.LastMessage.Content)
				s.Equal(2, conv.UnreadCount)
			},
		},
		{
			// conv has no LastMsgID → MsgRepo.GetByIDs must NOT be called.
			name: "success_no_last_msg_id",
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(existingConv(), nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(0), nil).Once()
				// no msgRepo expectation → mockery fails if GetByIDs is called unexpectedly
			},
			assertResult: func(s *conversationQuerySuite, conv *chat.Conversation) {
				s.Nil(conv.LastMessage)
			},
		},
		{
			// LastMessage fetch error is swallowed; conversation still returned.
			name: "last_message_fetch_error_swallowed",
			setupMock: func(deps testDeps) {
				conv := existingConv()
				conv.LastMsgID = "01MSG"
				deps.convRepo.EXPECT().
					GetByID(mock.Anything, convID).
					Return(conv, nil).Once()
				deps.unreadStore.EXPECT().
					Get(mock.Anything, userID, convID).
					Return(int64(0), nil).Once()
				deps.msgRepo.EXPECT().
					GetByIDs(mock.Anything, []string{"01MSG"}).
					Return(nil, assert.AnError).Once()
			},
			assertResult: func(s *conversationQuerySuite, conv *chat.Conversation) {
				s.NotNil(conv) // conv still returned despite msg fetch error
				s.Nil(conv.LastMessage)
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

	// makeConvs creates n conversations with decreasing UpdatedAt (index 0 = newest).
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

	// makeConvsWithMsg creates n conversations each referencing a last message.
	makeConvsWithMsg := func(n int) []chat.Conversation {
		convs := makeConvs(n)
		for i := range convs {
			convs[i].LastMsgID = fmt.Sprintf("01MSG%02d", i)
		}
		return convs
	}

	type testDeps struct {
		convRepo    *chatmocks.MockConversationRepository
		msgRepo     *chatmocks.MockMessageRepository
		unreadStore *chatmocks.MockUnreadStore
	}

	newDeps := func(t *testing.T) testDeps {
		return testDeps{
			convRepo:    chatmocks.NewMockConversationRepository(t),
			msgRepo:     chatmocks.NewMockMessageRepository(t),
			unreadStore: chatmocks.NewMockUnreadStore(t),
		}
	}

	newUC := func(d testDeps) *usecase.ConversationQueryUseCase {
		return usecase.NewQueryUseCase(usecase.QueryDeps{
			ConvRepo: d.convRepo,
			MsgRepo:  d.msgRepo,
			Unread:   d.unreadStore,
		})
	}

	type testCase struct {
		name         string
		params       chat.GetConversationsParams
		setupMock    func(deps testDeps)
		wantErr      error
		assertResult func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string])
	}

	defaultParams := chat.GetConversationsParams{
		UserID: userID,
		Query:  common.CursorQueryParams[string]{Limit: 10},
	}

	tests := []testCase{
		{
			name:   "repo_error",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "unread_error",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(2), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// GetList returns 2 convs, limit=10 → no next page.
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
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 2)
				s.False(result.HasNextPage)
				s.Empty(result.NextCursor)
			},
		},
		{
			// GetList returns limit+1=11 → HasNextPage=true, data trimmed to 10.
			// NextCursor = GetCursor() of the 10th item (index 9).
			name:   "success_has_next_page",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(11), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 10)
				s.True(result.HasNextPage)
				s.Equal(result.Data[9].GetCursor(), result.NextCursor)
			},
		},
		{
			// UnreadCount populated correctly from Redis map.
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
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Equal(7, result.Data[0].UnreadCount)
			},
		},
		{
			// State="" → usecase defaults to StateActive before calling repo.
			name:   "success_state_defaults_to_active",
			params: defaultParams,
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
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 1)
			},
		},
		{
			// Convs with LastMsgID → bulk fetch via GetByIDs → LastMessage populated.
			name:   "success_populates_last_messages",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				convs := makeConvsWithMsg(2)
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(convs, nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
				// Bulk fetch: both last-msg IDs collected in a single call.
				deps.msgRepo.EXPECT().
					GetByIDs(mock.Anything, mock.MatchedBy(func(ids []string) bool {
						return len(ids) == 2
					})).
					Return([]chat.Message{
						{ID: "01MSG00", Content: "first"},
						{ID: "01MSG01", Content: "second"},
					}, nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 2)
				for _, c := range result.Data {
					s.NotNil(c.LastMessage)
				}
			},
		},
		{
			// Convs with no LastMsgID → GetByIDs must NOT be called.
			name:   "success_no_last_msg_ids_skips_fetch",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(makeConvs(2), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
				// no msgRepo expectation — if GetByIDs is called the mock will fail
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				for _, c := range result.Data {
					s.Nil(c.LastMessage)
				}
			},
		},
		{
			// Bulk last-message fetch fails → error swallowed, convs still returned without LastMessage.
			name:   "last_message_bulk_fetch_error_swallowed",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				convs := makeConvsWithMsg(1)
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return(convs, nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
				deps.msgRepo.EXPECT().
					GetByIDs(mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 1)
				s.Nil(result.Data[0].LastMessage)
			},
		},
		{
			// Cursor pagination: explicit state=pending filter passes through.
			name: "success_pending_state_filter",
			params: chat.GetConversationsParams{
				UserID: userID,
				State:  chat.StatePending,
				Query:  common.CursorQueryParams[string]{Limit: 5},
			},
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.MatchedBy(func(p chat.GetConversationsParams) bool {
						return p.State == chat.StatePending
					})).
					Return(makeConvs(3), nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Len(result.Data, 3)
			},
		},
		{
			// Empty result: repo returns no convs → empty paginated result with no next page.
			name:   "success_empty_inbox",
			params: defaultParams,
			setupMock: func(deps testDeps) {
				deps.convRepo.EXPECT().
					GetList(mock.Anything, mock.Anything).
					Return([]chat.Conversation{}, nil).Once()
				deps.unreadStore.EXPECT().
					GetAll(mock.Anything, userID).
					Return(map[string]int64{}, nil).Once()
			},
			assertResult: func(s *conversationQuerySuite, result common.CursorPaginatedResult[chat.Conversation, string]) {
				s.Empty(result.Data)
				s.False(result.HasNextPage)
				s.Empty(result.NextCursor)
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
