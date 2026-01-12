package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type UserHandler struct {
	users service.UserService
}

func NewUserHandler(user service.UserService) *UserHandler {
	return &UserHandler{
		users: user,
	}
}

// Profile godoc
//
//	@Summary		Get user profile
//	@Description	Get current user profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	domain.UserResponse
//	@Router			/users/me [get]
func (h *UserHandler) Profile(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), payload.UserID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, user)
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update user profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.UpdateProfileRequest	true	"Update Profile Request"
//	@Success		200		{object}	domain.UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Router			/users/me [patch]
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

	user, err := h.users.UpdateProfile(c.Request.Context(), payload.UserID, &req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, user)

}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change user password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.ChangePasswordRequest	true	"Change Password Request"
//	@Success		200		{string}	string							"password changed successfully"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Router			/users/password [put]
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

	if err := h.users.ChangePassword(c.Request.Context(), payload.UserID, &req); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, "password changed successfully")
}

// UpdateAvatar godoc
//
//	@Summary		Update avatar
//	@Description	Upload and update user avatar image
//	@Tags			User
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			avatar	formData	file	true	"Avatar file"
//	@Success		200		{object}	domain.UserResponse
//	@Router			/users/avatar [post]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		pkg.BadRequest(c, "File avatar is required")
		return
	}

	if fileHeader.Size > 5*1024*1024 {
		pkg.BadRequest(c, "File to large (Max 5MB)")
		return
	}

	url, err := h.users.UpdateAvatar(c.Request.Context(), payload.UserID, fileHeader)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, gin.H{
		"message": "Avatar update successfully",
		"url":     url,
	})

}
