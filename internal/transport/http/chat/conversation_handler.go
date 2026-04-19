package chat

import (
	"github.com/gin-gonic/gin"

	chatdomain "air-social/internal/domain/chat"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/shared"
	"air-social/pkg"
)

type ConversationHandler struct {
	uc chatdomain.ConversationUseCase
}

func NewConversationHandler(uc chatdomain.ConversationUseCase) ConversationHandler {
	return ConversationHandler{uc: uc}
}

// CreateDirect godoc
//
//	@Summary		Create or get direct conversation
//	@Description	Creates a new direct (1-on-1) conversation between the authenticated user and the target user.
//	@Description	If a direct conversation between them already exists, the existing one is returned.
//	@Description	The recipient's initial state is "pending" unless they already follow the sender, in which case it is "active".
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateDirectReq	true	"Target user"
//	@Success		200		{object}	ConversationRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/direct [post]
func (h ConversationHandler) CreateDirect(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateDirectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if claims.UserID == req.TargetUserID {
		pkg.BadRequest(c, "cannot create conversation with yourself")
		return
	}

	conv, err := h.uc.Write.CreateOrGetDirect(c.Request.Context(), claims.UserID, req.TargetUserID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toConversationResponse(conv))
}

// CreateGroup godoc
//
//	@Summary		Create group conversation
//	@Description	Creates a new group conversation with the authenticated user as admin.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateGroupReq	true	"Group details"
//	@Success		200		{object}	ConversationRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/group [post]
func (h ConversationHandler) CreateGroup(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	conv, err := h.uc.Write.CreateGroup(c.Request.Context(), chatdomain.CreateGroupParams{
		CreatorID:    claims.UserID,
		MemberIDs:    req.ParticipantIDs,
		Name:         req.Name,
		AvatarKey:    req.AvatarKey,
		ClientConvID: req.ClientConvID,
	})
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toConversationResponse(conv))
}

// UpdateGroup godoc
//
//	@Summary		Update group conversation
//	@Description	Updates a group conversation's name and/or avatar. Only admins can update group info.
//	@Description	At least one field (name or avatar_key) must be provided.
//	@Tags			Chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string			true	"Conversation ID (ULID)"
//	@Param			body	body		UpdateGroupReq	true	"Fields to update (at least one required)"
//	@Success		200		{object}	ConversationRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations/{id} [patch]
func (h ConversationHandler) UpdateGroup(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req UpdateGroupReq
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	conv, err := h.uc.Write.UpdateGroup(c.Request.Context(), chatdomain.UpdateGroupParams{
		ConvID:    c.Param("id"),
		ActorID:   claims.UserID,
		Name:      req.Name,
		AvatarKey: req.AvatarKey,
	})
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toConversationResponse(conv))
}

// GetConversation godoc
//
//	@Summary		Get conversation by ID
//	@Description	Returns a single conversation the authenticated user is a participant of.
//	@Description	Includes unread count from Redis. LastMessage is populated once message layer is implemented.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Conversation ID (ULID)"
//	@Success		200	{object}	ConversationRes
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/conversations/{id} [get]
func (h ConversationHandler) GetConversation(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req PathIDParam
	if err := c.ShouldBindUri(&req); err != nil {
		pkg.BadRequest(c, "invalid conversation id")
		return
	}

	res, err := h.uc.Query.GetConversation(c.Request.Context(), req.ID, claims.UserID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toConversationResponse(res))
}

// GetConversations godoc
//
//	@Summary		List conversations
//	@Description	Returns cursor-paginated conversation inbox for the authenticated user.
//	@Description	Default state is "active". Use ?state=pending to fetch message requests.
//	@Description	Cursor is the next_cursor value from the previous response — pass it back as-is.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			state	query		string	false	"Participant state: active (default) | pending | ignored | muted"
//	@Param			cursor	query		string	false	"Pagination cursor from previous response"
//	@Param			limit	query		int		false	"Page size (default: 10, max: 50)"
//	@Success		200		{object}	ConversationsRes
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/conversations [get]
func (h ConversationHandler) GetConversations(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req GetConversationsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.uc.Query.GetConversations(c.Request.Context(), req.ToDomain(claims.UserID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	data := make([]ConversationRes, len(result.Data))
	for i := range result.Data {
		data[i] = toConversationResponse(&result.Data[i])
	}
	pkg.Success(c, ConversationsRes{
		Data: data,
		Meta: shared.MetaCursor[string]{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	})
}
