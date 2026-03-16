package like

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/like"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  like.UseCase
}

func NewHandler(provider common.LinkProvider, u like.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  u,
	}
}

// LikePost godoc
//
//	@Summary		Like a post
//	@Description	Like a post
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id}/likes [post]
func (h Handler) LikePost(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	if err := h.usecase.LikePost(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// GetPostLikers godoc
//
//	@Summary		Get users who liked a post
//	@Description	Get a list of users who liked a specific post using cursor-based pagination.
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Post ID"
//	@Param			cursor	query		int		false	"Cursor for pagination (last like ID)"
//	@Param			limit	query		int		false	"Number of items to return"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[LikerResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts/{id}/likes [get]
func (h Handler) GetPostLikers(c *gin.Context) {
	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	var req CursorQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.GetPostLikers(c.Request.Context(), req.ToDomain(path.ID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toLikerListResponse(result))
}

// GetCommentLikers godoc
//
//	@Summary		Get users who liked a comment
//	@Description	Get a list of users who liked a specific comment using cursor-based pagination.
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Comment ID"
//	@Param			cursor	query		int		false	"Cursor for pagination (last like ID)"
//	@Param			limit	query		int		false	"Number of items to return"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[LikerResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/comments/{id}/likes [get]
func (h Handler) GetCommentLikers(c *gin.Context) {
	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid comment id")
		return
	}

	var req CursorQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.GetCommentLikers(c.Request.Context(), req.ToDomain(path.ID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toLikerListResponse(result))
}

// UnlikePost godoc
//
//	@Summary		Unlike a post
//	@Description	Unlike a post
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id}/likes [delete]
func (h Handler) UnlikePost(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	if err := h.usecase.UnlikePost(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// LikeComment godoc
//
//	@Summary		Like a comment
//	@Description	Like a comment
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Comment ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/comments/{id}/likes [post]
func (h Handler) LikeComment(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid comment id")
		return
	}

	if err := h.usecase.LikeComment(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// UnlikeComment godoc
//
//	@Summary		Unlike a comment
//	@Description	Unlike a comment
//	@Tags			Like
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Comment ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/comments/{id}/likes [delete]
func (h Handler) UnlikeComment(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid comment id")
		return
	}

	if err := h.usecase.UnlikeComment(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

func (h Handler) toLikerResponse(l like.Liker) LikerResponse {
	return LikerResponse{
		ID:         l.UserID,
		FullName:   l.FullName,
		Avatar:     h.provider.PublicFile(l.Avatar),
		IsVerified: l.IsVerified,
	}
}

func (h Handler) toLikerListResponse(result common.CursorPaginatedResult[like.Liker, int64]) CursorPaginatedResponse[LikerResponse] {
	data := make([]LikerResponse, len(result.Data))
	for i, v := range result.Data {
		data[i] = h.toLikerResponse(v)
	}

	return CursorPaginatedResponse[LikerResponse]{
		Data: data,
		Meta: MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}
