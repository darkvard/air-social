package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type UserHandler struct {
	userSvc service.UserService
	url     domain.URLFactory
}

func NewUserHandler(userSvc service.UserService, url domain.URLFactory) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		url:     url,
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
//	@Success		200	{object}	dto.UserResponse
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/me [get]
func (h *UserHandler) Profile(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	user, err := h.userSvc.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.Unauthorized(c, "account has been deleted or suspended")
			return
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.mapToResponse(user))
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update user profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UpdateProfileRequest	true	"Update Profile Request"
//	@Success		200		{object}	dto.UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		409		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req dto.UpdateProfileRequest
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.UpdateProfileParams{
		UserID:   claims.UserID,
		FullName: req.FullName,
		Bio:      req.Bio,
		Location: req.Location,
		Website:  req.Website,
		Username: req.Username,
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.Unauthorized(c, "account has been deleted or suspended")
			return
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.mapToResponse(user))

}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change user password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.ChangePasswordRequest	true	"Change Password Request"
//	@Success		200		{string}	string						"password changed successfully"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.ChangePasswordParams{
		UserID:          claims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.userSvc.ChangePassword(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "password changed successfully", nil)
}

// ConfirmFileUpload godoc
//
//	@Summary		Confirm file upload
//	@Description	Confirm that the file has been uploaded successfully and update the user profile with the new image URL.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.ConfirmProfileImageRequest	true	"Confirm Upload Request"
//	@Success		200		{object}	map[string]string				"Returns upload success message and public URL"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/profile-image/confirm [post]
func (h *UserHandler) ConfirmFileUpload(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req dto.ConfirmProfileImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.ConfirmFileParams{
		EntityID:  claims.UserID,
		ObjectKey: req.ObjectKey,
		Domain:    req.Domain,
		Feature:   req.Feature,
	}

	fileURL, err := h.userSvc.ConfirmImageUpload(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, gin.H{
		"message": "Profile image updated successfully",
		"url":     fileURL,
	})

}

func (h *UserHandler) mapToResponse(user *domain.User) dto.UserResponse {
	avatar := h.url.PublicFileURL(user.Profile.Avatar)
	cover := h.url.PublicFileURL(user.Profile.CoverImage)
	return dto.NewUserResponse(user, avatar, cover)
}
