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

// todo:
//  - handler: remove package dto, handler (old)
//	- usecase: add unit tests	
//	- service: remove services (old)
//	- domain: remove files...
//	- test: test all api
//
