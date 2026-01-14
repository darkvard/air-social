package handler

import (
	"path/filepath"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type UserHandler struct {
	srv service.UserService
}

func NewUserHandler(srv service.UserService) *UserHandler {
	return &UserHandler{
		srv: srv,
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

	user, err := h.srv.GetByID(c.Request.Context(), payload.UserID)
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

	user, err := h.srv.UpdateProfile(c.Request.Context(), payload.UserID, &req)
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

	if err := h.srv.ChangePassword(c.Request.Context(), payload.UserID, &req); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, "password changed successfully")
}

// PresignedFileUpload godoc
//
//	@Summary		Get presigned upload URL
//	@Description	Generate a presigned URL for uploading a file (avatar or cover) to object storage.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.PresignedFileUploadRequest	true	"Presigned Upload Request"
//	@Success		200		{object}	domain.PresignedFileResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Router			/users/file/presigned [post]
func (h *UserHandler) PresignedFileUpload(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.PresignedFileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if v := pkg.ValidateImageFile(req.FileName); v != nil {
		pkg.HandleValidationResult(c, v)
		return
	}

	res, err := h.srv.PresignedImageUpload(
		c.Request.Context(),
		domain.PresignedFile{
			UserID: payload.UserID,
			Ext:    filepath.Ext(req.FileName),
			Typ:    domain.FileType(req.FileType),
		},
	)

	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, res)
}

// ConfirmFileUpload godoc
//
//	@Summary		Confirm file upload
//	@Description	Confirm that the file has been uploaded successfully and update the user profile with the new image URL.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.ConfirmFileUploadRequest	true	"Confirm Upload Request"
//	@Success		200		{object}	map[string]string				"Returns upload success message and public URL"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Router			/users/file/confirm [post]
func (h *UserHandler) ConfirmFileUpload(c *gin.Context) {
	payload, err := middleware.GetAuthPayload(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req domain.ConfirmFileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	finalURL, err := h.srv.ConfirmImageUpload(c.Request.Context(), domain.ConfirmFile{
		UserID:     payload.UserID,
		ObjectName: req.ObjectName,
		Typ:        domain.FileType(req.FileType),
	})
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, gin.H{
		"message": "Upload success",
		"url":     finalURL,
	})

}
