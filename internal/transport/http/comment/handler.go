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
//	@Success		201		{object}	CreateResponse
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
	}

	if len(req.Media) > 0 {
		params.Media = make([]comment.Media, len(req.Media))
		for i, v := range req.Media {
			params.Media[i] = comment.Media{
				MediaKey:  v.MediaKey,
				MediaType: v.MediaType,
			}
		}
	}

	result, err := h.usecase.CreateComment(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.toCreateResponse(result))
}

func (h Handler) GetComments(c *gin.Context) {

}

func (h Handler) UpdateComment(c *gin.Context) {

}

func (h Handler) DeleteComment(c *gin.Context) {

}

func (h Handler) GetReplies(c *gin.Context) {

}

func (h Handler) toCreateResponse(c *comment.Comment) CreateResponse {
	if c == nil {
		return CreateResponse{}
	}

	resp := CreateResponse{
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
