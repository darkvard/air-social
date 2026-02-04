package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type FollowHandler struct {
	srv service.FollowService
}

func NewFollowHandler(srv service.FollowService) *FollowHandler {
	return &FollowHandler{
		srv: srv,
	}
}

// Follow godoc
//
//	@Summary		Follow a user
//	@Description	Follow another user.
//	@Tags			Follow
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int		true	"User ID"
//	@Success		200	{string}	string	"follow success"
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/{id}/follow [post]
func (h *FollowHandler) Follow(c *gin.Context) {
	var req dto.FollowRequest
	if err := c.ShouldBindUri(&req); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.srv.Follow(c.Request.Context(), claims.UserID, req.TargetUserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "follow success", nil)
}

// Unfollow godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow another user.
//	@Tags			Follow
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int		true	"User ID"
//	@Success		200	{string}	string	"unfollow success"
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/{id}/follow [delete]
func (h *FollowHandler) Unfollow(c *gin.Context) {
	var req dto.FollowRequest
	if err := c.ShouldBindUri(&req); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.srv.Unfollow(c.Request.Context(), claims.UserID, req.TargetUserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "unfollow success", nil)
}

func (h *FollowHandler) GetFollowers(c *gin.Context) {

}

func (h *FollowHandler) GetFollowings(c *gin.Context) {

}
