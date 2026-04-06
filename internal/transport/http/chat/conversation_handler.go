package chat

import (
	"github.com/gin-gonic/gin"

	chatdomain "air-social/internal/domain/chat"
	"air-social/internal/transport/http/middleware"
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
//	@Success		200		{object}	ConversationResponse
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
//	@Success		200		{object}	ConversationResponse
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

// GetConversation godoc
//
//	@Summary		Get conversation by ID
//	@Description	Returns a single conversation the authenticated user is a participant of.
//	@Description	Includes unread count from Redis. LastMessage is populated once message layer is implemented.
//	@Tags			Chat
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Conversation ID (ULID)"
//	@Success		200	{object}	ConversationResponse
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

	var req GetConversationReq
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

func (h ConversationHandler) GetConversations(c *gin.Context) {

}
