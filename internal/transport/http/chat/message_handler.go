package chat

import (
	"github.com/gin-gonic/gin"

	chatdomain "air-social/internal/domain/chat"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/shared"
	"air-social/pkg"
)

type MessageHandler struct {
	uc chatdomain.Messenger
}

func NewMessageHandler(uc chatdomain.Messenger) MessageHandler {
	return MessageHandler{uc: uc}
}

// GetMessages godoc
//
//	@Summary		List messages in a conversation
//	@Description	Returns cursor-paginated messages for a conversation the authenticated user is a participant of.
//	@Description	Messages are returned newest-first. Pass next_cursor from the previous response to paginate.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Conversation ID (ULID)"
//	@Param			cursor	query		string	false	"Pagination cursor from previous response"
//	@Param			limit	query		int		false	"Page size (default: 20, max: 50)"
//	@Success		200		{object}	MessagesRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages [get]
func (h MessageHandler) GetMessages(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param PathIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid conversation id")
		return
	}

	var req GetMessagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.uc.GetMessages(c.Request.Context(), req.ToDomain(param.ID, claims.UserID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	data := make([]MessageRes, len(result.Data))
	for i := range result.Data {
		data[i] = toMessageRes(&result.Data[i])
	}
	pkg.Success(c, MessagesRes{
		Data: data,
		Meta: shared.MetaCursor[string]{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	})
}

// SendMessage godoc
//
//	@Summary		Send a message
//	@Description	Sends a message to a conversation. The authenticated user must be an active or pending participant.
//	@Description	Pending senders are implicitly accepted on first message. Supports idempotency via client_msg_id.
//	@Description	For non-text types (image, video, audio, file), content must be a valid MinIO object key.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Conversation ID (ULID)"
//	@Param			body	body		SendMessageReq	true	"Message payload"
//	@Success		201		{object}	MessageRes		"New message created"
//	@Success		200		{object}	MessageRes		"Duplicate message (idempotent retry via client_msg_id)"
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages [post]
func (h MessageHandler) SendMessage(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param PathIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid conversation id")
		return
	}

	var req SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	msg, isNew, err := h.uc.SendMessage(c.Request.Context(), req.ToDomain(param.ID, claims.UserID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	if isNew {
		pkg.Created(c, toMessageRes(msg))
	} else {
		pkg.Success(c, toMessageRes(msg))
	}
}

// EditMessage godoc
//
//	@Summary		Edit a message
//	@Description	Updates the content of a message. Only the original sender can edit. Deleted messages cannot be edited.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Conversation ID (ULID)"
//	@Param			msgID	path		string			true	"Message ID (ULID)"
//	@Param			body	body		EditMessageReq	true	"New content"
//	@Success		200		{object}	MessageRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages/{msgID} [patch]
func (h MessageHandler) EditMessage(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param MessagePathParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid path params")
		return
	}

	var req EditMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	msg, err := h.uc.EditMessage(c.Request.Context(), chatdomain.EditMessageParams{
		MessageID: param.MsgID,
		UserID:    claims.UserID,
		Content:   req.Content,
	})
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toMessageRes(msg))
}

// DeleteMessage godoc
//
//	@Summary		Delete a message
//	@Description	Soft-deletes a message. Only the original sender can delete. Content is cleared but the message row is preserved for thread context.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Conversation ID (ULID)"
//	@Param			msgID	path		string	true	"Message ID (ULID)"
//	@Success		204		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages/{msgID} [delete]
func (h MessageHandler) DeleteMessage(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param MessagePathParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid path params")
		return
	}

	if err := h.uc.DeleteMessage(c.Request.Context(), param.MsgID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// AddReaction godoc
//
//	@Summary		Add a reaction to a message
//	@Description	Adds an emoji reaction from the authenticated user to the specified message.
//	@Description	Allowed reactions: like | love | haha | wow | sad | angry
//	@Description	Deleted messages cannot be reacted to.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string		true	"Conversation ID (ULID)"
//	@Param			msgID	path		string		true	"Message ID (ULID)"
//	@Param			body	body		ReactionReq	true	"Emoji reaction"
//	@Success		204		{object}	pkg.Response
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages/{msgID}/reactions [post]
func (h MessageHandler) AddReaction(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param MessagePathParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid path params")
		return
	}

	var req ReactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if err := h.uc.AddReaction(c.Request.Context(), param.MsgID, claims.UserID, chatdomain.ReactionType(req.Type)); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// RemoveReaction godoc
//
//	@Summary		Remove a reaction from a message
//	@Description	Removes the authenticated user's reaction from the specified message.
//	@Description	Deleted messages cannot be modified.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Conversation ID (ULID)"
//	@Param			msgID	path		string	true	"Message ID (ULID)"
//	@Success		204		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id}/messages/{msgID}/reactions [delete]
func (h MessageHandler) RemoveReaction(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param MessagePathParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid path params")
		return
	}

	if err := h.uc.RemoveReaction(c.Request.Context(), param.MsgID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// MarkRead godoc
//
//	@Summary		Mark conversation as read
//	@Description	Resets unread count to 0 and updates the authenticated user's last-read pointer to the conversation's latest message.
//	@Description	No request body needed. The server derives the latest message from the conversation state.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Conversation ID (ULID)"
//	@Success		204	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/conversations/{id}/read [patch]
func (h MessageHandler) MarkRead(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param PathIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid conversation id")
		return
	}

	if err := h.uc.MarkRead(c.Request.Context(), param.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// MarkUnread godoc
//
//	@Summary		Mark conversation as unread
//	@Description	Sets the unread count to 1 for the authenticated user. Private reminder only — does NOT roll back
//	@Description	the seen status visible to other participants. Once a message is seen, it remains seen for the sender.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Conversation ID (ULID)"
//	@Success		204	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/conversations/{id}/read [delete]
func (h MessageHandler) MarkUnread(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var param PathIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		pkg.BadRequest(c, "invalid conversation id")
		return
	}

	if err := h.uc.MarkUnread(c.Request.Context(), param.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}
