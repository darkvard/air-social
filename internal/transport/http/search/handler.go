package search

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	"air-social/internal/domain/search"
	"air-social/internal/transport/http/middleware"
	postdto "air-social/internal/transport/http/post"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  search.UseCase
}

func NewHandler(provider common.LinkProvider, usecase search.UseCase) Handler {
	return Handler{provider: provider, usecase: usecase}
}

// SearchUsers godoc
//
//	@Summary		Search users
//	@Description	Search for users by username or full name using cursor-based pagination. Requires authentication.
//	@Tags			Search
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			query	query		string	true	"Search keyword matched against username and full_name (case-insensitive)"
//	@Param			cursor	query		int		false	"Cursor for pagination (ID of the last item from the previous page, 0 for first page)"
//	@Param			limit	query		int		false	"Number of items to return (default=10, max=50)"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[UserResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/search/users [get]
func (h Handler) SearchUsers(c *gin.Context) {
	var req CursorQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.SearchUsers(c.Request.Context(), req.toUserParams())
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toUserListResponse(result))
}

// SearchPosts godoc
//
//	@Summary		Search posts
//	@Description	Search public posts by content using full-text search with cursor-based pagination. Requires authentication.
//	@Tags			Search
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			query	query		string	true	"Search keyword matched against post content"
//	@Param			cursor	query		int		false	"Cursor for pagination (ID of the last item from the previous page, 0 for first page)"
//	@Param			limit	query		int		false	"Number of items to return (default=10, max=50)"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[PostResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/search/posts [get]
func (h Handler) SearchPosts(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req CursorQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.SearchPosts(c.Request.Context(), req.toPostParams(claims.UserID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPostListResponse(result))
}

func (h Handler) toUserListResponse(result common.CursorPaginatedResult[search.User, int64]) CursorPaginatedResponse[UserResponse] {
	data := make([]UserResponse, len(result.Data))
	for i, v := range result.Data {
		data[i] = UserResponse{
			ID:       v.ID,
			Username: v.Username,
			FullName: v.Fullname,
			Avatar:   h.provider.PublicFile(v.Avatar),
			Verified: v.Verified,
		}
	}
	return CursorPaginatedResponse[UserResponse]{
		Data: data,
		Meta: MetaCursor{NextCursor: result.NextCursor, HasNextPage: result.HasNextPage},
	}
}

func (h Handler) toPostListResponse(result common.CursorPaginatedResult[*post.Post, int64]) CursorPaginatedResponse[PostResponse] {
	data := make([]PostResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = postdto.ToPostResponse(p, h.provider.PublicFile)
	}
	return CursorPaginatedResponse[PostResponse]{
		Data: data,
		Meta: MetaCursor{NextCursor: result.NextCursor, HasNextPage: result.HasNextPage},
	}
}
