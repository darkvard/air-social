package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type UserHandler struct {
	user service.UserService
}

func NewUserHandler(user service.UserService) *UserHandler {
	return &UserHandler{
		user: user,
	}
}

func (h *UserHandler) Profile(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	user, err := h.user.GetByID(c.Request.Context(), payload.UserID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	user, err := h.user.UpdateProfile(c.Request.Context(), payload.UserID, &req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, user)

}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if req.NewPassword == req.CurrentPassword {
		pkg.BadRequest(c, "new password must be different from current password")
		return
	}

	if err := h.user.ChangePassword(c.Request.Context(), payload.UserID, &req); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, "password changed successfully")
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {

}
