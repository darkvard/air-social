package like

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/like"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	usecase like.UseCase
}

func NewHandler(u like.UseCase) Handler {
	return Handler{usecase: u}
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
