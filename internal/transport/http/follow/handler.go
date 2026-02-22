package follow

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  follow.UseCase
}

func NewHandler(provider common.LinkProvider, usecase follow.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
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
//	@Param			id	path		int	true	"User ID"
//	@Success		201	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/{id}/follow [post]
func (h Handler) Follow(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	if err := h.usecase.Follow(c.Request.Context(), claims.UserID, path.ID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.Created(c, nil)
}

// Unfollow godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow another user.
//	@Tags			Follow
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"User ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/users/{id}/follow [delete]
func (h Handler) Unfollow(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	if err := h.usecase.Unfollow(c.Request.Context(), claims.UserID, path.ID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.NoContent(c)
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
//	@Param			page	query		int		false	"Page number"		default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			sort	query		string	false	"Sort order"		Enums(latest, oldest, name_asc, name_desc)
//	@Success		200		{object}	common.OffsetPaginatedResult[UserFollowResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/followers [get]
func (h Handler) GetFollowers(c *gin.Context) {
	params, err := buildGetFollowsParams(c)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	result, err := h.usecase.GetFollowers(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, common.NewOffsetPaginatedResult(
		h.toUserFollowResponse(result.Data),
		result.Total,
		result.Page,
		result.Limit,
	))
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
//	@Param			page	query		int		false	"Page number"		default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			sort	query		string	false	"Sort order"		Enums(latest, oldest, name_asc, name_desc)
//	@Success		200		{object}	common.OffsetPaginatedResult[UserFollowResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/followings [get]
func (h Handler) GetFollowings(c *gin.Context) {
	params, err := buildGetFollowsParams(c)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	result, err := h.usecase.GetFollowings(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, common.NewOffsetPaginatedResult(
		h.toUserFollowResponse(result.Data),
		result.Total,
		result.Page,
		result.Limit,
	))
}

func (h Handler) toUserFollowResponse(user []follow.FollowUser) []UserFollowResponse {
	res := make([]UserFollowResponse, len(user))
	for i, u := range user {
		res[i] = UserFollowResponse{
			ID:          u.ID,
			FullName:    u.FullName,
			Avatar:      h.provider.PublicFile(u.Avatar),
			IsVerified:  u.IsVerified,
			IsFollowing: u.IsFollowing,
			IsFollower:  u.IsFollower,
		}
	}
	return res
}

func buildGetFollowsParams(c *gin.Context) (follow.GetFollowsParams, error) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		return follow.GetFollowsParams{}, pkg.NewError(pkg.ErrUnauthorized, err.Error())
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		return follow.GetFollowsParams{}, pkg.NewError(pkg.ErrBadRequest, "invalid user id")
	}

	var query QueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		return follow.GetFollowsParams{}, pkg.NewError(pkg.ErrBadRequest, "invalid query parameters")
	}

	return follow.GetFollowsParams{
		ViewerID: claims.UserID,
		TargetID: path.ID,
		Paging:   query.ToDomain(),
	}, nil
}
