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
	url       domain.URLFactory
}

func NewFollowHandler(followSvc service.FollowService, userSvc service.UserService, url domain.URLFactory) *FollowHandler {
	return &FollowHandler{
		followSvc: followSvc,
		userSvc:   userSvc,
		url:       url,
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

// GetFollowers godoc
//
//	@Summary		Get followers
//	@Description	Get a paginated list of users who follow the specified user
//	@Tags			Follow
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"User ID"
//	@Param			page	query		int		false	"Page number"
//	@Param			limit	query		int		false	"Items per page"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest, name_asc, name_desc)
//	@Success		200		{object}	dto.PaginatedResponse[dto.UserFollowResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/followers [get]
func (h *FollowHandler) GetFollowers(c *gin.Context) {
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

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		pkg.BadRequest(c, "invalid query")
		return
	}

	params := domain.FollowParams{
		QueryParams:   query.ToDomain(),
		TargetUserID:  path.ID,
		CurrentUserID: claims.UserID,
	}

	result, err := h.followSvc.GetFollowers(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	response := dto.NewPaginatedResponse(domain.PaginatedResult[dto.UserFollowResponse]{
		Data:       h.mapToFollowResponses(result.Data),
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	})

	pkg.Success(c, response)
}

// GetFollowings godoc
//
//	@Summary		Get followings
//	@Description	Get a paginated list of users that the specified user is following
//	@Tags			Follow
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"User ID"
//	@Param			page	query		int		false	"Page number"
//	@Param			limit	query		int		false	"Items per page"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest, name_asc, name_desc)
//	@Success		200		{object}	dto.PaginatedResponse[dto.UserFollowResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/followings [get]
func (h *FollowHandler) GetFollowings(c *gin.Context) {
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

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		pkg.BadRequest(c, "invalid query")
		return
	}

	params := domain.FollowParams{
		QueryParams:   query.ToDomain(),
		TargetUserID:  path.ID,
		CurrentUserID: claims.UserID,
	}

	result, err := h.followSvc.GetFollowings(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	response := dto.NewPaginatedResponse(domain.PaginatedResult[dto.UserFollowResponse]{
		Data:       h.mapToFollowResponses(result.Data),
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	})

	pkg.Success(c, response)
}

// Internal helper

func (h *FollowHandler) mapToFollowResponses(users []domain.SocialUser) []dto.UserFollowResponse {
	res := make([]dto.UserFollowResponse, len(users))
	for i, u := range users {
		res[i] = dto.UserFollowResponse{
			ID:           u.User.ID,
			Username:     u.User.Username,
			FullName:     u.User.Profile.FullName,
			Avatar:       h.url.PublicFileURL(u.User.Profile.Avatar),
			IsVerified:   u.User.Status.Verified,
			IsFollowing:  u.Relation.IsFollowing,
			IsFollowedBy: u.Relation.IsFollowedBy,
		}
	}
	return res
}
