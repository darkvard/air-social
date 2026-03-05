package comment

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/comment"
	"air-social/internal/domain/common"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  comment.UseCase
}

func NewHandler(provider common.LinkProvider, usecase comment.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}

// CreateComment godoc
//
//	@Summary		Create a new comment
//	@Description	Create a new comment on a post or reply to an existing comment.
//	@Tags			Comment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"Post ID"
//	@Param			request	body		CreateRequest	true	"Create Comment Request"
//	@Success		201		{object}	CommentResponse
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts/{id}/comments [post]
func (h Handler) CreateComment(c *gin.Context) {
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

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := comment.CreateParams{
		UserID:   claims.UserID,
		PostID:   path.ID,
		Content:  req.Content,
		ParentID: req.ParentID,
		Media:    toMediaParams(req.Media),
	}

	result, err := h.usecase.CreateComment(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.toCommentResponse(result))
}

func (h Handler) GetComments(c *gin.Context) {

}

// UpdateComment godoc
//
//	@Summary		Update a comment
//	@Description	Update a comment's content and/or media. Only the author can update their comment.
//	@Tags			Comment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"Comment ID"
//	@Param			request	body		UpdateRequest	true	"Update Comment Request"
//	@Success		200		{object}	CommentResponse
//	@Failure		400		{object}	pkg.Response
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/comments/{id} [patch]
func (h Handler) UpdateComment(c *gin.Context) {
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

	var req UpdateRequest
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := comment.UpdateParams{
		CommentID: path.ID,
		UserID:    claims.UserID,
		Content:   req.Content,
	}
	if req.Media != nil {
		params.Media = toMediaParams(req.Media)
	}

	updatedComment, err := h.usecase.UpdateComment(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toCommentResponse(updatedComment))
}

// DeleteComment godoc
//
//	@Summary		Delete a comment
//	@Description	Delete a comment by ID. Only the author can delete their comment.
//	@Tags			Comment
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Comment ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/comments/{id} [delete]
func (h Handler) DeleteComment(c *gin.Context) {
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

	if err := h.usecase.DeleteComment(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

func (h Handler) GetReplies(c *gin.Context) {

}

func (h Handler) toCommentResponse(c *comment.Comment) CommentResponse {
	if c == nil {
		return CommentResponse{}
	}

	resp := CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		ParentID:  c.ParentID,
		CreatedAt: c.CreatedAt,
	}

	if len(c.Media) > 0 {
		resp.Media = make([]MediaItemResponse, len(c.Media))
		for i, m := range c.Media {
			resp.Media[i] = MediaItemResponse{
				URL:       h.provider.PublicFile(m.MediaKey),
				MediaType: m.MediaType,
			}
		}
	}
	return resp
}

func toMediaParams(data []MediaItemInput) []comment.Media {
	result := make([]comment.Media, len(data))
	if data == nil {
		return result
	}
	for i, v := range data {
		result[i] = comment.Media{
			MediaKey:  v.MediaKey,
			MediaType: v.MediaType,
		}
	}
	return result
}
