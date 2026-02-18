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
	userSvc    service.UserService
	urlFactory domain.URLFactory
}

func NewUserHandler(userSvc service.UserService, urlFactory domain.URLFactory) *UserHandler {
	return &UserHandler{
		userSvc:    userSvc,
		urlFactory: urlFactory,
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
//	@Success		200	{object}	dto.UserDetailResponse
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/me [get]
func (h *UserHandler) Profile(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	user, err := h.userSvc.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.Unauthorized(c, "account has been deleted or suspended")
			return
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserResponse(user))
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
//	@Success		200		{object}	dto.UserDetailResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		409		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
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

	pkg.Success(c, h.toUserResponse(user))

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
	claims, err := middleware.GetTokenClaims(c)
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
//	@Param			request	body		dto.BulkConfirmProfileImageRequest	true	"Confirm Upload Request List"
//	@Success		200		{array}		dto.ConfirmFileResponse				"Returns list of confirmed files"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/profile-image/confirm [post]
func (h *UserHandler) ConfirmFileUpload(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req dto.BulkConfirmProfileImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := h.toConfirmParams(claims.UserID, req.Files)

	results, err := h.userSvc.ConfirmImageUpload(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toConfirmResponse(results))
}

// Internal helper

func (h *UserHandler) toUserResponse(user *domain.User) dto.UserDetailResponse {
	return dto.NewUserDetailResponse(
		user,
		h.urlFactory.PublicFileURL(user.Profile.Avatar),
		h.urlFactory.PublicFileURL(user.Profile.CoverImage),
	)
}

func (h *UserHandler) toConfirmParams(userID int64, files []dto.ConfirmProfileImageRequest) []domain.ConfirmFileParams {
	params := make([]domain.ConfirmFileParams, len(files))
	for i, req := range files {
		params[i] = domain.ConfirmFileParams{
			EntityID:  userID,
			ObjectKey: req.ObjectKey,
			Domain:    req.Domain,
			Feature:   req.Feature,
		}
	}
	return params
}

func (h *UserHandler) toConfirmResponse(results []domain.ConfirmFileResult) []dto.ConfirmFileResponse {
	resp := make([]dto.ConfirmFileResponse, len(results))
	for i, res := range results {
		resp[i] = dto.ConfirmFileResponse{
			Domain:  res.Domain,
			Feature: res.Feature,
			URL:     res.URL,
		}
	}
	return resp
}
