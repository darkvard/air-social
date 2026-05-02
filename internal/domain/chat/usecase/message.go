package usecase

import (
	"context"
	"errors"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/pkg"
)

type MessageDeps struct {
	MsgRepo       chat.MessageRepository
	ConvRepo      chat.ConversationRepository
	Event         common.EventPublisher
	Hub           RealtimePublisher // WebSocket hub: delivers real-time events to connected clients
	Unread        chat.UnreadStore
	MediaVerifier MediaVerifier
}

type MessageUseCase struct {
	deps MessageDeps
}

func NewMessageUseCase(d MessageDeps) *MessageUseCase {
	return &MessageUseCase{deps: d}
}

func (u *MessageUseCase) GetMessages(ctx context.Context, params chat.GetMessagesParams) (common.CursorPaginatedResult[chat.Message, string], error) {
	var empty common.CursorPaginatedResult[chat.Message, string]
	params.Query.NormalizePagination()

	conv, err := u.deps.ConvRepo.GetByID(ctx, params.ConversationID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}
	if conv == nil {
		return empty, pkg.ErrNotFound
	}
	if conv.FindParticipant(params.UserID) == nil {
		return empty, pkg.ErrForbidden
	}

	msgs, err := u.deps.MsgRepo.GetList(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return common.NewCursorPaginatedResult(msgs, params.Query.Limit), nil
}

func (u *MessageUseCase) SendMessage(ctx context.Context, params chat.SendMessageParams) (*chat.Message, bool, error) {
	conv, err := u.deps.ConvRepo.GetByID(ctx, params.ConversationID)
	if err != nil {
		return nil, false, pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil, false, pkg.ErrNotFound
	}

	sender, err := u.validateSender(conv, params.SenderID)
	if err != nil {
		return nil, false, err
	}

	if err := u.verifyMedia(ctx, params.Type, params.Content); err != nil {
		return nil, false, err
	}

	if err := u.validateReplyTo(ctx, params.ReplyToID, params.ConversationID); err != nil {
		return nil, false, err
	}

	u.implicitlyAccept(ctx, conv.ID, sender)

	msg, isNew, err := u.persistWithIdempotency(ctx, u.buildMessage(params))
	if err != nil {
		return nil, false, err
	}

	if isNew {
		u.dispatchAfterSend(ctx, msg, conv)
	}

	return msg, isNew, nil
}

func (u *MessageUseCase) EditMessage(ctx context.Context, params chat.EditMessageParams) (*chat.Message, error) {
	msg, err := u.deps.MsgRepo.GetByID(ctx, params.MessageID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if msg == nil {
		return nil, pkg.ErrNotFound
	}
	if msg.SenderID != params.UserID {
		return nil, pkg.ErrForbidden
	}
	if msg.IsDeleted {
		return nil, pkg.NewError(pkg.ErrBadRequest, "cannot edit a deleted message")
	}

	msg.Content = params.Content
	msg.IsEdited = true

	if err := u.deps.MsgRepo.Update(ctx, msg); err != nil {
		return nil, pkg.OrInternalError(err)
	}

	_ = u.deps.Hub.PublishMessageEdited(ctx, msg)
	return msg, nil
}

func (u *MessageUseCase) DeleteMessage(ctx context.Context, msgID string, userID int64) error {
	msg, err := u.deps.MsgRepo.GetByID(ctx, msgID)
	if err != nil {
		return pkg.OrInternalError(err)
	}
	if msg == nil {
		return pkg.ErrNotFound
	}
	if msg.SenderID != userID {
		return pkg.ErrForbidden
	}
	if msg.IsDeleted {
		return pkg.NewError(pkg.ErrBadRequest, "message already deleted")
	}

	if err := u.deps.MsgRepo.SoftDelete(ctx, msgID); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = u.deps.Hub.PublishMessageDeleted(ctx, msgID, msg.ConversationID)

	return nil
}

func (u *MessageUseCase) AddReaction(ctx context.Context, msgID string, userID int64, reaction chat.ReactionType) error {
	if !reaction.IsValid() {
		return pkg.NewError(pkg.ErrBadRequest, "invalid reaction type")
	}

	msg, conv, err := u.loadMsgAndCheckParticipant(ctx, msgID, userID)
	if err != nil {
		return err
	}
	if msg.IsDeleted {
		return pkg.NewError(pkg.ErrBadRequest, "cannot react to a deleted message")
	}

	r := &chat.Reaction{
		UserID:    userID,
		Type:      reaction,
		CreatedAt: pkg.TimeNowUTC(),
	}

	if err := u.deps.MsgRepo.AddReaction(ctx, msgID, r); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = u.deps.Hub.PublishReactionAdded(ctx, msgID, conv.ID, *r)

	return nil
}

func (u *MessageUseCase) RemoveReaction(ctx context.Context, msgID string, userID int64) error {
	msg, conv, err := u.loadMsgAndCheckParticipant(ctx, msgID, userID)
	if err != nil {
		return err
	}
	if msg.IsDeleted {
		return pkg.NewError(pkg.ErrBadRequest, "cannot remove reaction from a deleted message")
	}

	if err := u.deps.MsgRepo.RemoveReaction(ctx, msgID, userID); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = u.deps.Hub.PublishReactionRemoved(ctx, msgID, conv.ID, userID)

	return nil
}

func (u *MessageUseCase) MarkRead(ctx context.Context, convID string, userID int64) error {
	conv, err := u.loadConvAndCheckParticipant(ctx, convID, userID)
	if err != nil {
		return err
	}

	if conv.LastMsgID != "" {
		if err := u.deps.ConvRepo.UpdateLastRead(ctx, convID, userID, conv.LastMsgID); err != nil {
			return pkg.OrInternalError(err)
		}
	}

	_ = u.deps.Unread.Reset(ctx, userID, convID)
	_ = u.deps.Hub.PublishReadEvent(ctx, convID, userID, conv.LastMsgID)

	return nil
}

func (u *MessageUseCase) MarkUnread(ctx context.Context, convID string, userID int64) error {
	if _, err := u.loadConvAndCheckParticipant(ctx, convID, userID); err != nil {
		return err
	}

	if err := u.deps.Unread.Reset(ctx, userID, convID); err != nil {
		return pkg.OrInternalError(err)
	}

	if err := u.deps.Unread.Increment(ctx, userID, convID); err != nil {
		return pkg.OrInternalError(err)
	}

	return nil
}

// loadConvAndCheckParticipant loads a conversation and verifies the user is an active participant.
// Returns ErrNotFound if the conversation doesn't exist, ErrForbidden if not a participant or ignored.
func (u *MessageUseCase) loadConvAndCheckParticipant(ctx context.Context, convID string, userID int64) (*chat.Conversation, error) {
	conv, err := u.deps.ConvRepo.GetByID(ctx, convID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil, pkg.ErrNotFound
	}
	if _, err := u.validateSender(conv, userID); err != nil {
		return nil, err
	}
	return conv, nil
}

// loadMsgAndCheckParticipant loads a message and verifies the user is a participant of its conversation.
func (u *MessageUseCase) loadMsgAndCheckParticipant(ctx context.Context, msgID string, userID int64) (*chat.Message, *chat.Conversation, error) {
	msg, err := u.deps.MsgRepo.GetByID(ctx, msgID)
	if err != nil {
		return nil, nil, pkg.OrInternalError(err)
	}
	if msg == nil {
		return nil, nil, pkg.ErrNotFound
	}
	conv, err := u.loadConvAndCheckParticipant(ctx, msg.ConversationID, userID)
	if err != nil {
		return nil, nil, err
	}
	return msg, conv, nil
}

// validateSender checks that the sender is a participant who is allowed to send.
// StateIgnored senders are blocked — they explicitly declined the conversation;
// they must accept it first via AcceptConversation before participating.
func (u *MessageUseCase) validateSender(conv *chat.Conversation, senderID int64) (*chat.Participant, error) {
	sender := conv.FindParticipant(senderID)
	if sender == nil || sender.State == chat.StateIgnored {
		return nil, pkg.ErrForbidden
	}
	return sender, nil
}

// verifyMedia checks that non-text messages reference a valid object in storage.
func (u *MessageUseCase) verifyMedia(ctx context.Context, msgType chat.MessageType, content string) error {
	if msgType == chat.MessageText {
		return nil
	}
	if err := u.deps.MediaVerifier.VerifyMedia(ctx, []string{content}); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.NewError(pkg.ErrBadRequest, "media not found or invalid")
		}
		return pkg.OrInternalError(err)
	}
	return nil
}

// implicitlyAccept promotes a pending sender to active when they first reply.
// This mirrors the behaviour of accepting a message request by replying.
// Errors are ignored — the message still goes through even if the state update fails.
func (u *MessageUseCase) implicitlyAccept(ctx context.Context, convID string, sender *chat.Participant) {
	if sender.State == chat.StatePending {
		_ = u.deps.ConvRepo.UpdateParticipantState(ctx, convID, sender.UserID, chat.StateActive)
	}
}

// buildMessage constructs the domain Message from send params.
func (u *MessageUseCase) buildMessage(params chat.SendMessageParams) *chat.Message {
	now := pkg.TimeNowUTC()
	return &chat.Message{
		ID:             pkg.NewULID(),
		ConversationID: params.ConversationID,
		SenderID:       params.SenderID,
		Type:           params.Type,
		Content:        params.Content,
		ClientMsgID:    params.ClientMsgID,
		ReplyToID:      params.ReplyToID,
		Reactions:      []chat.Reaction{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// persistWithIdempotency saves the message and handles duplicate-key conflicts.
// Returns (msg, isNew=true) on fresh create.
// When a conflict occurs and the request carries a ClientMsgID we treat it as a
// client retry: look up the already-persisted message and return (existing, isNew=false).
// Side effects (Touch, Unread, RT publish) must only fire when isNew=true.
// Without a ClientMsgID the caller opted out of idempotency — the conflict surfaces as ErrInternal.
func (u *MessageUseCase) persistWithIdempotency(ctx context.Context, msg *chat.Message) (*chat.Message, bool, error) {
	if err := u.deps.MsgRepo.Create(ctx, msg); err != nil {
		if errors.Is(err, pkg.ErrConflict) && msg.ClientMsgID != "" {
			existing, findErr := u.deps.MsgRepo.GetByClientMsgID(ctx, msg.ClientMsgID, msg.SenderID)
			if findErr != nil {
				return nil, false, pkg.OrInternalError(findErr)
			}
			if existing != nil {
				return existing, false, nil
			}
		}
		return nil, false, pkg.OrInternalError(err)
	}
	return msg, true, nil
}

// validateReplyTo checks that reply_to_id, when provided, refers to a message that exists
// and belongs to the same conversation. Soft-deleted messages are intentionally allowed —
// the UI should render them as "This message was deleted" in the reply preview.
func (u *MessageUseCase) validateReplyTo(ctx context.Context, replyToID *string, convID string) error {
	if replyToID == nil {
		return nil
	}
	parent, err := u.deps.MsgRepo.GetByID(ctx, *replyToID)
	if err != nil {
		return pkg.OrInternalError(err)
	}
	if parent == nil {
		return pkg.NewError(pkg.ErrBadRequest, "reply_to message not found")
	}
	if parent.ConversationID != convID {
		return pkg.NewError(pkg.ErrBadRequest, "reply_to message does not belong to this conversation")
	}
	return nil
}

// dispatchAfterSend fires all fire-and-forget side effects after a message is persisted:
// updating the conversation's last-message pointer, incrementing unread counts for other
// participants, and publishing real-time + async events.
func (u *MessageUseCase) dispatchAfterSend(ctx context.Context, msg *chat.Message, conv *chat.Conversation) {
	_ = u.deps.ConvRepo.Touch(ctx, conv.ID, msg.ID)

	recipientIDs := make([]int64, 0, len(conv.Participants)-1)
	for i := range conv.Participants {
		p := &conv.Participants[i]
		if p.UserID != msg.SenderID {
			_ = u.deps.Unread.Increment(ctx, p.UserID, conv.ID)
			recipientIDs = append(recipientIDs, p.UserID)
		}
	}

	_ = u.deps.Hub.PublishNewMessage(ctx, msg, conv)
	_ = u.deps.Event.Publish(ctx, common.NewEvent(common.EventMessageCreated, common.MessageEventPayload{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		RecipientIDs:   recipientIDs,
	}))
}
