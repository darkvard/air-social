package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type FollowHandler struct {
	followSvc service.FollowService
	userSvc   service.UserService
}

func NewFollowHandler(followSvc service.FollowService, userSvc service.UserService) *FollowHandler {
	return &FollowHandler{
		followSvc: followSvc,
		userSvc:   userSvc,
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
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path dto.FollowPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	if err := h.followSvc.Follow(c.Request.Context(), claims.UserID, path.ID); err != nil {
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
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path dto.FollowPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	if err := h.followSvc.Unfollow(c.Request.Context(), claims.UserID, path.ID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "unfollow success", nil)
}

func (h *FollowHandler) GetFollowers(c *gin.Context) {
	var path dto.FollowPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		pkg.BadRequest(c, "invalid query")
		return
	}

	result, err := h.followSvc.GetFollowers(c.Request.Context(), domain.FollowParams{
		UserID: path.ID,
		Page:   query.Page,
		Limit:  query.Limit,
	})
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, dto.PaginatedResult{
		Data:  h.mapToUserResponses(result.Users),
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
	})

	// todo: unit test
	// todo: FormatPublicURL (media, user, auth, here) ->  URLFactory.FormatPublicURL(::key) 
}

func (h *FollowHandler) GetFollowings(c *gin.Context) {

}

// Internal helper

func (h *FollowHandler) mapToUserResponses(users []domain.User) []dto.UserResponse {
	responses := make([]dto.UserResponse, len(users))

	for i := range users {
		u := &users[i]
		avatarURL := h.userSvc.FormatPublicURL(u.Profile.Avatar)
		coverURL := h.userSvc.FormatPublicURL(u.Profile.CoverImage)
		responses[i] = dto.NewUserResponse(u, avatarURL, coverURL)
	}

	return responses
}
