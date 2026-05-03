package user

import (
	"errors"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/media"
	"air-social/internal/domain/user"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  user.UseCase
}

func NewHandler(provider common.LinkProvider, usecase user.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
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
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/me [get]
func (h Handler) Profile(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	user, err := h.usecase.Fetch.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.Unauthorized(c, "account has been deleted or suspended")
			return
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserResponse(*user))
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update user profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpdateProfileRequest	true	"Update Profile Request"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me [patch]
func (h Handler) UpdateProfile(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req UpdateProfileRequest
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := user.UpdateParams{
		UserID:   claims.UserID,
		FullName: req.FullName,
		Bio:      req.Bio,
		Location: req.Location,
		Website:  req.Website,
		Username: req.Username,
	}

	user, err := h.usecase.Profile.UpdateProfile(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			pkg.Unauthorized(c, "account has been deleted or suspended")
			return
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserResponse(*user))
}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change user password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		ChangePasswordRequest	true	"Change Password Request"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/password [put]
func (h Handler) ChangePassword(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := user.ChangePasswordParams{
		UserID:          claims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := h.usecase.Account.ChangePassword(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// UpdateAvatar godoc
//
//	@Summary		Update avatar
//	@Description	Confirm avatar upload and update user profile
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpdateImageRequest	true	"Update Avatar Request"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/avatar [put]
func (h Handler) UpdateAvatar(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := media.ConfirmParams{
		EntityID:  claims.UserID,
		ObjectKey: req.ObjectKey,
		Domain:    media.DomainUser,
		Feature:   media.FeatureAvatar,
	}

	user, err := h.usecase.Profile.UpdateAvatar(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserResponse(*user))
}

// UpdateCover godoc
//
//	@Summary		Update cover image
//	@Description	Confirm cover image upload and update user profile
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpdateImageRequest	true	"Update Cover Request"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/me/cover [put]
func (h Handler) UpdateCover(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var req UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := media.ConfirmParams{
		EntityID:  claims.UserID,
		ObjectKey: req.ObjectKey,
		Domain:    media.DomainUser,
		Feature:   media.FeatureCover,
	}

	user, err := h.usecase.Profile.UpdateCover(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserResponse(*user))
}

func (h Handler) toUserResponse(user user.User) UserResponse {
	return UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Profile: ProfileResponse{
			FullName:   user.Profile.FullName,
			Bio:        user.Profile.Bio,
			Avatar:     h.provider.PublicFile(user.Profile.Avatar),
			CoverImage: h.provider.PublicFile(user.Profile.CoverImage),
			Location:   user.Profile.Location,
			Website:    user.Profile.Website,
		},
		Status: StatusResponse{
			Verified:   user.Status.Verified,
			VerifiedAt: user.Status.VerifiedAt,
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Version:   user.Version,
	}
}

// GetPresenceBatch godoc
//
//	@Summary		Get online status for multiple users
//	@Description	Returns online status for up to 100 users in a single request. Presence is refreshed automatically via WebSocket keep-alive every 20s and expires 30s after disconnect.
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		PresenceBatchRequest	true	"List of user IDs (max 100)"
//	@Success		200		{object}	[]PresenceResponse
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/presence [post]
func (h Handler) GetPresenceBatch(c *gin.Context) {
	var req PresenceBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}
	statuses, err := h.usecase.Presence.GetBatchStatus(c.Request.Context(), req.UserIDs)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	result := make([]PresenceResponse, 0, len(req.UserIDs))
	for _, id := range req.UserIDs {
		result = append(result, PresenceResponse{UserID: id, Online: statuses[id]})
	}
	pkg.Success(c, result)
}
