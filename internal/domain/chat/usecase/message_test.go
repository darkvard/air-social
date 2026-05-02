package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/chat"
	chatmocks "air-social/internal/domain/chat/mocks"
	"air-social/internal/domain/chat/usecase"
	ucmocks "air-social/internal/domain/chat/usecase/mocks"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/common"
	"air-social/pkg"
)

type messageSuite struct {
	suite.Suite
}

func TestMessageUseCaseSuite(t *testing.T) {
	suite.Run(t, new(messageSuite))
}

type msgTestDeps struct {
	msgRepo       *chatmocks.MockMessageRepository
	convRepo      *chatmocks.MockConversationRepository
	event         *commonmocks.MockEventPublisher
	rtPublisher   *ucmocks.MockRealtimePublisher
	unread        *chatmocks.MockUnreadStore
	mediaVerifier *ucmocks.MockMediaVerifier
}

func newMsgDeps(t interface {
	mock.TestingT
	Cleanup(func())
}) msgTestDeps {
	return msgTestDeps{
		msgRepo:       chatmocks.NewMockMessageRepository(t),
		convRepo:      chatmocks.NewMockConversationRepository(t),
		event:         commonmocks.NewMockEventPublisher(t),
		rtPublisher:   ucmocks.NewMockRealtimePublisher(t),
		unread:        chatmocks.NewMockUnreadStore(t),
		mediaVerifier: ucmocks.NewMockMediaVerifier(t),
	}
}

func (d msgTestDeps) newUC() *usecase.MessageUseCase {
	return usecase.NewMessageUseCase(usecase.MessageDeps{
		MsgRepo:       d.msgRepo,
		ConvRepo:      d.convRepo,
		Event:         d.event,
		Hub:           d.rtPublisher,
		Unread:        d.unread,
		MediaVerifier: d.mediaVerifier,
	})
}

// ── Shared fixtures ───────────────────────────────────────────────────────────

var (
	msgCtx    = context.Background()
	msgConvID = "01CONV"
	msgSender = int64(1)
	msgOther  = int64(2)
	msgThird  = int64(3)
)

func ptr(s string) *string { return &s }

// activeDirectConv returns a 2-participant direct conversation with both members active.
func activeDirectConv() *chat.Conversation {
	return &chat.Conversation{
		ID:   msgConvID,
		Type: chat.ConversationDirect,
		Participants: []chat.Participant{
			{UserID: msgSender, State: chat.StateActive},
			{UserID: msgOther, State: chat.StateActive},
		},
	}
}

// activeGroupConv returns a 3-participant group conversation.
func activeGroupConv() *chat.Conversation {
	return &chat.Conversation{
		ID:   msgConvID,
		Type: chat.ConversationGroup,
		Participants: []chat.Participant{
			{UserID: msgSender, State: chat.StateActive, Role: chat.RoleAdmin},
			{UserID: msgOther, State: chat.StateActive, Role: chat.RoleMember},
			{UserID: msgThird, State: chat.StateActive, Role: chat.RoleMember},
		},
	}
}

// ── TestSendMessage ────────────────────────────────────────────────────────────

func (s *messageSuite) TestSendMessage() {
	baseParams := chat.SendMessageParams{
		ConversationID: msgConvID,
		SenderID:       msgSender,
		Type:           chat.MessageText,
		Content:        "hello",
		ClientMsgID:    "client-abc",
	}

	pendingConv := &chat.Conversation{
		ID:   msgConvID,
		Type: chat.ConversationDirect,
		Participants: []chat.Participant{
			{UserID: msgSender, State: chat.StatePending},
			{UserID: msgOther, State: chat.StateActive},
		},
	}

	tests := []struct {
		name      string
		params    chat.SendMessageParams
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:   "conv_not_found",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "sender_not_participant",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(&chat.Conversation{
					ID:           msgConvID,
					Participants: []chat.Participant{{UserID: msgOther, State: chat.StateActive}},
				}, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "sender_ignored",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(&chat.Conversation{
					ID: msgConvID,
					Participants: []chat.Participant{
						{UserID: msgSender, State: chat.StateIgnored},
						{UserID: msgOther, State: chat.StateActive},
					},
				}, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "sender_muted_can_still_send",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				mutedConv := &chat.Conversation{
					ID:   msgConvID,
					Type: chat.ConversationDirect,
					Participants: []chat.Participant{
						{UserID: msgSender, State: chat.StateMuted},
						{UserID: msgOther, State: chat.StateActive},
					},
				}
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(mutedConv, nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "invalid_media",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageImage,
				Content:        "some-object-key",
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(msgCtx, []string{"some-object-key"}).Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "media_verify_internal_error",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageVideo,
				Content:        "video-key",
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.mediaVerifier.EXPECT().VerifyMedia(msgCtx, []string{"video-key"}).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "implicit_accept",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(pendingConv, nil).Once()
				d.convRepo.EXPECT().UpdateParticipantState(msgCtx, msgConvID, msgSender, chat.StateActive).Return(nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(nil).Once()
			},
		},
		{
			// Implicit accept fails → message still sent (fire-and-forget).
			name:   "implicit_accept_state_update_fails",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(pendingConv, nil).Once()
				d.convRepo.EXPECT().UpdateParticipantState(msgCtx, msgConvID, msgSender, chat.StateActive).Return(pkg.ErrInternal).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(nil).Once()
			},
		},
		{
			// Conflict + ClientMsgID → fetch existing → return without re-firing side effects.
			name:   "idempotent_retry",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				existing := &chat.Message{ID: "01EXISTING", ClientMsgID: "client-abc"}
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrConflict).Once()
				d.msgRepo.EXPECT().GetByClientMsgID(msgCtx, "client-abc", msgSender).Return(existing, nil).Once()
				// No Touch, Unread, RT, or Event expectations — side effects must NOT fire.
			},
		},
		{
			// Conflict + no ClientMsgID → no idempotency guarantee → ErrInternal.
			name: "conflict_without_client_msg_id",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageText,
				Content:        "hi",
				ClientMsgID:    "",
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrConflict).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// Conflict + ClientMsgID set, but GetByClientMsgID returns not-found (race resolved elsewhere).
			name:   "idempotent_retry_existing_not_found",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrConflict).Once()
				d.msgRepo.EXPECT().GetByClientMsgID(msgCtx, "client-abc", msgSender).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// Conflict + ClientMsgID set, GetByClientMsgID itself fails.
			name:   "idempotent_retry_fetch_error",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrConflict).Once()
				d.msgRepo.EXPECT().GetByClientMsgID(msgCtx, "client-abc", msgSender).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// Side effects fail → message still returned (fire-and-forget).
			name:   "success_side_effects_fail",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(pkg.ErrInternal).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(pkg.ErrInternal).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(pkg.ErrInternal).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(pkg.ErrInternal).Once()
			},
		},
		{
			// Group conv: unread incremented for each non-sender participant.
			name: "group_conv_increments_unread_for_all_others",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageText,
				Content:        "group message",
				ClientMsgID:    "client-grp",
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeGroupConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				// Both non-sender participants must get unread incremented.
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgThird, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.MatchedBy(func(e common.Event) bool {
					p, ok := e.Data.(common.MessageEventPayload)
					return ok && len(p.RecipientIDs) == 2
				})).Return(nil).Once()
			},
		},
		{
			// reply_to_id set, but the referenced message does not exist.
			name: "reply_to_not_found",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageText,
				Content:        "hi",
				ReplyToID:      ptr("01GHOST"),
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetByID(msgCtx, "01GHOST").Return(nil, nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			// reply_to_id points to a message in a different conversation.
			name: "reply_to_wrong_conversation",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageText,
				Content:        "hi",
				ReplyToID:      ptr("01FOREIGN"),
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetByID(msgCtx, "01FOREIGN").Return(&chat.Message{
					ID:             "01FOREIGN",
					ConversationID: "OTHER_CONV",
				}, nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			// reply_to_id pointing to a soft-deleted message is allowed.
			name: "reply_to_deleted_message_ok",
			params: chat.SendMessageParams{
				ConversationID: msgConvID,
				SenderID:       msgSender,
				Type:           chat.MessageText,
				Content:        "hi",
				ReplyToID:      ptr("01DELETED"),
				ClientMsgID:    "client-reply",
			},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetByID(msgCtx, "01DELETED").Return(&chat.Message{
					ID:             "01DELETED",
					ConversationID: msgConvID,
					IsDeleted:      true,
				}, nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:   "success",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(nil).Once()
				d.convRepo.EXPECT().Touch(msgCtx, msgConvID, mock.AnythingOfType("string")).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgOther, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishNewMessage(msgCtx, mock.Anything, mock.Anything).Return(nil).Once()
				d.event.EXPECT().Publish(msgCtx, mock.Anything).Return(nil).Once()
			},
		},
		{
			name: "create_error",
			params: chat.SendMessageParams{ConversationID: msgConvID, SenderID: msgSender, Type: chat.MessageText, Content: "hi"},
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().Create(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			msg, _, err := uc.SendMessage(msgCtx, tt.params)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(msg)
			} else {
				s.NoError(err)
				s.NotNil(msg)
			}
		})
	}
}

// ── TestGetMessages ────────────────────────────────────────────────────────────

func (s *messageSuite) TestGetMessages() {
	baseParams := chat.GetMessagesParams{
		ConversationID: msgConvID,
		UserID:         msgSender,
		Query:          common.CursorQueryParams[string]{Limit: 20},
	}

	makeMsgs := func(n int) []chat.Message {
		msgs := make([]chat.Message, n)
		for i := range msgs {
			msgs[i] = chat.Message{ID: pkg.NewULID(), ConversationID: msgConvID}
		}
		return msgs
	}

	tests := []struct {
		name         string
		params       chat.GetMessagesParams
		setupMock    func(d msgTestDeps)
		wantErr      error
		assertResult func(s *messageSuite, r common.CursorPaginatedResult[chat.Message, string])
	}{
		{
			name:   "conv_not_found",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "conv_repo_error",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "user_not_participant",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(&chat.Conversation{
					ID:           msgConvID,
					Participants: []chat.Participant{{UserID: msgOther, State: chat.StateActive}},
				}, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			// Ignored/muted participants still see messages — valid participant is valid participant.
			name:   "ignored_participant_can_read_messages",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(&chat.Conversation{
					ID: msgConvID,
					Participants: []chat.Participant{
						{UserID: msgSender, State: chat.StateIgnored},
						{UserID: msgOther, State: chat.StateActive},
					},
				}, nil).Once()
				d.msgRepo.EXPECT().GetList(msgCtx, mock.Anything).Return(makeMsgs(2), nil).Once()
			},
			assertResult: func(s *messageSuite, r common.CursorPaginatedResult[chat.Message, string]) {
				s.Len(r.Data, 2)
			},
		},
		{
			name:   "msg_repo_error",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetList(msgCtx, mock.Anything).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// No cursor → first page, returns all messages.
			name:   "success_first_page",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetList(msgCtx, mock.Anything).Return(makeMsgs(5), nil).Once()
			},
			assertResult: func(s *messageSuite, r common.CursorPaginatedResult[chat.Message, string]) {
				s.Len(r.Data, 5)
				s.False(r.HasNextPage)
			},
		},
		{
			// Fetch limit+1 returns exactly limit+1 → HasNextPage=true.
			name:   "success_has_next_page",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetList(msgCtx, mock.Anything).Return(makeMsgs(21), nil).Once()
			},
			assertResult: func(s *messageSuite, r common.CursorPaginatedResult[chat.Message, string]) {
				s.Len(r.Data, 20)
				s.True(r.HasNextPage)
				s.Equal(r.Data[19].GetCursor(), r.NextCursor)
			},
		},
		{
			name:   "success_empty_conversation",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().GetList(msgCtx, mock.Anything).Return([]chat.Message{}, nil).Once()
			},
			assertResult: func(s *messageSuite, r common.CursorPaginatedResult[chat.Message, string]) {
				s.Empty(r.Data)
				s.False(r.HasNextPage)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			result, err := uc.GetMessages(msgCtx, tt.params)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
				if tt.assertResult != nil {
					tt.assertResult(s, result)
				}
			}
		})
	}
}

// ── TestEditMessage ────────────────────────────────────────────────────────────

func (s *messageSuite) TestEditMessage() {
	const msgID = "01MSG"

	baseParams := chat.EditMessageParams{
		MessageID: msgID,
		UserID:    msgSender,
		Content:   "updated content",
	}

	existingMsg := func() *chat.Message {
		return &chat.Message{
			ID:             msgID,
			ConversationID: msgConvID,
			SenderID:       msgSender,
			Type:           chat.MessageText,
			Content:        "original",
		}
	}

	tests := []struct {
		name         string
		params       chat.EditMessageParams
		setupMock    func(d msgTestDeps)
		wantErr      error
		assertResult func(s *messageSuite, m *chat.Message)
	}{
		{
			name:   "msg_not_found",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "msg_repo_get_error",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "not_sender",
			params: chat.EditMessageParams{MessageID: msgID, UserID: msgOther, Content: "hacked"},
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			// Deleted messages cannot be edited.
			name:   "msg_already_deleted",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				msg := existingMsg()
				msg.IsDeleted = true
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(msg, nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:   "repo_update_error",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.msgRepo.EXPECT().Update(msgCtx, mock.AnythingOfType("*chat.Message")).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			params: baseParams,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.msgRepo.EXPECT().Update(msgCtx, mock.MatchedBy(func(m *chat.Message) bool {
					return m.Content == "updated content" && m.IsEdited
				})).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishMessageEdited(msgCtx, mock.Anything).Return(nil).Once()
			},
			assertResult: func(s *messageSuite, m *chat.Message) {
				s.Equal("updated content", m.Content)
				s.True(m.IsEdited)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			msg, err := uc.EditMessage(msgCtx, tt.params)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(msg)
			} else {
				s.NoError(err)
				s.NotNil(msg)
				if tt.assertResult != nil {
					tt.assertResult(s, msg)
				}
			}
		})
	}
}

// ── TestDeleteMessage ──────────────────────────────────────────────────────────

func (s *messageSuite) TestDeleteMessage() {
	const msgID = "01MSG"

	existingMsg := func() *chat.Message {
		return &chat.Message{
			ID:             msgID,
			ConversationID: msgConvID,
			SenderID:       msgSender,
		}
	}

	tests := []struct {
		name      string
		msgID     string
		userID    int64
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:   "msg_not_found",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "msg_repo_get_error",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "not_sender",
			msgID:  msgID,
			userID: msgOther,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "soft_delete_repo_error",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.msgRepo.EXPECT().SoftDelete(msgCtx, msgID).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.msgRepo.EXPECT().SoftDelete(msgCtx, msgID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishMessageDeleted(msgCtx, msgID, msgConvID).Return(nil).Once()
			},
		},
		{
			// RT publish failure must not block the delete response (fire-and-forget).
			name:   "success_rt_publish_fails",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.msgRepo.EXPECT().SoftDelete(msgCtx, msgID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishMessageDeleted(msgCtx, msgID, msgConvID).Return(pkg.ErrInternal).Once()
			},
		},
		{
			// SoftDelete must clear reactions — verified at repo level; usecase just calls SoftDelete.
			name:   "soft_delete_clears_reactions",
			msgID:  msgID,
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				msg := existingMsg()
				msg.Reactions = []chat.Reaction{
					{UserID: msgOther, Type: chat.ReactionLike},
				}
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(msg, nil).Once()
				d.msgRepo.EXPECT().SoftDelete(msgCtx, msgID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishMessageDeleted(msgCtx, msgID, msgConvID).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			err := uc.DeleteMessage(msgCtx, tt.msgID, tt.userID)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── TestAddReaction ────────────────────────────────────────────────────────────

func (s *messageSuite) TestAddReaction() {
	const msgID = "01MSG"

	existingMsg := func() *chat.Message {
		return &chat.Message{
			ID:             msgID,
			ConversationID: msgConvID,
			SenderID:       msgSender,
		}
	}

	tests := []struct {
		name      string
		userID    int64
		reaction  chat.ReactionType
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:     "invalid_reaction",
			userID:   msgSender,
			reaction: "clown",
			setupMock: func(d msgTestDeps) {
				// validation fires before any repo call
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:     "msg_not_found",
			userID:   msgSender,
			reaction: chat.ReactionLike,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:     "msg_repo_get_error",
			userID:   msgSender,
			reaction: chat.ReactionLike,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:     "deleted_message_forbidden",
			userID:   msgSender,
			reaction: chat.ReactionLike,
			setupMock: func(d msgTestDeps) {
				deleted := existingMsg()
				deleted.IsDeleted = true
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(deleted, nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:     "non_participant_forbidden",
			userID:   msgThird, // not in the direct conv (only msgSender + msgOther)
			reaction: chat.ReactionLike,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:     "add_reaction_repo_error",
			userID:   msgSender,
			reaction: chat.ReactionLove,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().AddReaction(msgCtx, msgID, mock.AnythingOfType("*chat.Reaction")).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:     "success",
			userID:   msgSender,
			reaction: chat.ReactionLike,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().AddReaction(msgCtx, msgID, mock.MatchedBy(func(r *chat.Reaction) bool {
					return r.UserID == msgSender && r.Type == chat.ReactionLike
				})).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionAdded(msgCtx, msgID, msgConvID, mock.Anything).Return(nil).Once()
			},
		},
		{
			// RT publish failure must not block the response (fire-and-forget).
			name:     "success_rt_publish_fails",
			userID:   msgSender,
			reaction: chat.ReactionHaha,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().AddReaction(msgCtx, msgID, mock.Anything).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionAdded(msgCtx, msgID, msgConvID, mock.Anything).Return(pkg.ErrInternal).Once()
			},
		},
		{
			// Replace-or-add: user sends a different type — treated as a new AddReaction call at usecase level.
			name:     "success_all_valid_types",
			userID:   msgOther, // participant B can also react
			reaction: chat.ReactionAngry,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().AddReaction(msgCtx, msgID, mock.MatchedBy(func(r *chat.Reaction) bool {
					return r.Type == chat.ReactionAngry
				})).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionAdded(msgCtx, msgID, msgConvID, mock.Anything).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			err := uc.AddReaction(msgCtx, msgID, tt.userID, tt.reaction)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── TestRemoveReaction ─────────────────────────────────────────────────────────

func (s *messageSuite) TestRemoveReaction() {
	const msgID = "01MSG"

	existingMsg := func() *chat.Message {
		return &chat.Message{
			ID:             msgID,
			ConversationID: msgConvID,
			SenderID:       msgSender,
		}
	}

	tests := []struct {
		name      string
		userID    int64
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:   "msg_not_found",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "msg_repo_get_error",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "deleted_message_forbidden",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				deleted := existingMsg()
				deleted.IsDeleted = true
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(deleted, nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:   "non_participant_forbidden",
			userID: msgThird,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "remove_reaction_repo_error",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().RemoveReaction(msgCtx, msgID, msgSender).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().RemoveReaction(msgCtx, msgID, msgSender).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionRemoved(msgCtx, msgID, msgConvID, msgSender).Return(nil).Once()
			},
		},
		{
			// RT publish failure must not block the response (fire-and-forget).
			name:   "success_rt_publish_fails",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().RemoveReaction(msgCtx, msgID, msgSender).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionRemoved(msgCtx, msgID, msgConvID, msgSender).Return(pkg.ErrInternal).Once()
			},
		},
		{
			// Removing a non-existent reaction is a no-op at repo level (MongoDB $pull on missing).
			name:   "success_reaction_not_present_is_noop",
			userID: msgOther,
			setupMock: func(d msgTestDeps) {
				d.msgRepo.EXPECT().GetByID(msgCtx, msgID).Return(existingMsg(), nil).Once()
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.msgRepo.EXPECT().RemoveReaction(msgCtx, msgID, msgOther).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReactionRemoved(msgCtx, msgID, msgConvID, msgOther).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			err := uc.RemoveReaction(msgCtx, msgID, tt.userID)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── TestMarkRead ──────────────────────────────────────────────────────────────

func (s *messageSuite) TestMarkRead() {
	const lastMsgID = "01LASTMSG"

	// Conversation with a last message (normal case).
	convWithLastMsg := func() *chat.Conversation {
		c := activeDirectConv()
		c.LastMsgID = lastMsgID
		return c
	}

	tests := []struct {
		name      string
		userID    int64
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:   "conv_not_found",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "non_participant_forbidden",
			userID: msgThird,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			// UpdateLastRead failure blocks the response (not fire-and-forget).
			name:   "update_last_read_repo_error",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(convWithLastMsg(), nil).Once()
				d.convRepo.EXPECT().UpdateLastRead(msgCtx, msgConvID, msgSender, lastMsgID).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// Conversation has no messages yet — skip UpdateLastRead, still reset unread.
			name:   "conv_no_last_msg",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				// No UpdateLastRead expectation.
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReadEvent(msgCtx, msgConvID, msgSender, "").Return(nil).Once()
			},
		},
		{
			name:   "success",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(convWithLastMsg(), nil).Once()
				d.convRepo.EXPECT().UpdateLastRead(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReadEvent(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
			},
		},
		{
			// Unread reset failure must not block response (fire-and-forget).
			name:   "success_unread_reset_fails",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(convWithLastMsg(), nil).Once()
				d.convRepo.EXPECT().UpdateLastRead(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(pkg.ErrInternal).Once()
				d.rtPublisher.EXPECT().PublishReadEvent(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
			},
		},
		{
			// RT publish failure must not block response (fire-and-forget).
			name:   "success_rt_publish_fails",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(convWithLastMsg(), nil).Once()
				d.convRepo.EXPECT().UpdateLastRead(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(nil).Once()
				d.rtPublisher.EXPECT().PublishReadEvent(msgCtx, msgConvID, msgSender, lastMsgID).Return(pkg.ErrInternal).Once()
			},
		},
		{
			// Both side effects fail — still returns nil.
			name:   "success_all_side_effects_fail",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(convWithLastMsg(), nil).Once()
				d.convRepo.EXPECT().UpdateLastRead(msgCtx, msgConvID, msgSender, lastMsgID).Return(nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(pkg.ErrInternal).Once()
				d.rtPublisher.EXPECT().PublishReadEvent(msgCtx, msgConvID, msgSender, lastMsgID).Return(pkg.ErrInternal).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			err := uc.MarkRead(msgCtx, msgConvID, tt.userID)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ── TestMarkUnread ────────────────────────────────────────────────────────────

func (s *messageSuite) TestMarkUnread() {
	tests := []struct {
		name      string
		userID    int64
		setupMock func(d msgTestDeps)
		wantErr   error
	}{
		{
			name:   "conv_not_found",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(nil, nil).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "non_participant_forbidden",
			userID: msgThird,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			// Redis Reset fails → return error (Redis IS the operation here).
			name:   "reset_error",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			// Redis Increment fails → return error.
			name:   "increment_error",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgSender, msgConvID).Return(pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			userID: msgSender,
			setupMock: func(d msgTestDeps) {
				d.convRepo.EXPECT().GetByID(msgCtx, msgConvID).Return(activeDirectConv(), nil).Once()
				d.unread.EXPECT().Reset(msgCtx, msgSender, msgConvID).Return(nil).Once()
				d.unread.EXPECT().Increment(msgCtx, msgSender, msgConvID).Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newMsgDeps(s.T())
			tt.setupMock(d)
			uc := d.newUC()

			err := uc.MarkUnread(msgCtx, msgConvID, tt.userID)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
