package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type PostHandler struct {
	srv        service.PostService
	urlFactory domain.URLFactory
}

func NewPostHandler(srv service.PostService, urlFactory domain.URLFactory) *PostHandler {
	return &PostHandler{
		srv:        srv,
		urlFactory: urlFactory,
	}
}

// CreatePost godoc
//
//	@Summary		Create a new post
//	@Description	Create a new post with content and optional media attachments.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreatePostRequest	true	"Create Post Request"
//	@Success		201		{object}	dto.PostResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.CreatePostParams{
		UserID:     claims.UserID,
		Content:    req.Content,
		Visibility: domain.PostVisibility(req.Visibility),
	}

	if len(req.Media) > 0 {
		params.Media = make([]domain.PostMediaParams, len(req.Media))
		for i, v := range req.Media {
			params.Media[i] = h.toPostMediaParams(v)
		}
	}

	post, err := h.srv.CreatePost(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.toPostResponse(post))
}

func (h *PostHandler) GetPost(c *gin.Context) {

}

func (h *PostHandler) GetUserPosts(c *gin.Context) {

}

func (h *PostHandler) UpdatePost(c *gin.Context) {

}

func (h *PostHandler) DeletePost(c *gin.Context) {

}

// Internal helper

func (h *PostHandler) toPostMediaParams(req dto.MediaItemInput) domain.PostMediaParams {
	return domain.PostMediaParams{
		MediaKey:  req.MediaKey,
		MediaType: req.MediaType,
		Width:     req.Width,
		Height:    req.Height,
		Duration:  req.Duration,
		Size:      req.Size,
		FileName:  req.FileName,
	}
}

func (h *PostHandler) toPostResponse(post *domain.Post) dto.PostResponse {
	mediaResp := make([]dto.MediaItemResponse, len(post.Media))
	for i, m := range post.Media {
		mediaResp[i] = dto.MediaItemResponse{
			ID:        m.ID,
			URL:       h.urlFactory.PublicFileURL(m.MediaKey),
			MediaType: m.MediaType,
			Width:     m.Metadata.Width,
			Height:    m.Metadata.Height,
			Duration:  m.Metadata.Duration,
			FileName:  m.Metadata.FileName,
		}
	}

	return dto.PostResponse{
		ID:         post.ID,
		Content:    post.Content,
		Visibility: string(post.Visibility),
		Version:    post.Version,
		CreatedAt:  post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  post.UpdatedAt.Format(time.RFC3339),
		User: dto.UserCompactResponse{
			ID:       post.User.ID,
			Fullname: post.User.FullName,
			Avatar:   h.urlFactory.PublicFileURL(post.User.Avatar),
		},
		Media: mediaResp,
	}
}
