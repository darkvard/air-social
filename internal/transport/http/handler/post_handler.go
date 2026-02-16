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

// GetPost godoc
//
//	@Summary		Get post detail
//	@Description	Get post details by ID
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	dto.PostResponse
//	@Failure		400	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	var path dto.IDPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	post, err := h.srv.GetPostDetail(c.Request.Context(), path.ID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPostResponse(post))
}

// GetUserPosts godoc
//
//	@Summary		Get user posts
//	@Description	Get a list of posts for a specific user using cursor-based pagination.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int	true	"User ID"
//	@Param			cursor	query		int	false	"Cursor for pagination (last post ID)"
//	@Param			limit	query		int	false	"Number of items to return"
//	@Success		200		{object}	dto.CursorPaginatedResponse[dto.PostResponse]
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/posts [get]
func (h *PostHandler) GetUserPosts(c *gin.Context) {
	var path dto.IDPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	var req dto.CursorPaginationQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.srv.GetUserPosts(c.Request.Context(), path.ID, req.ToDomain())
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, dto.NewCursorPaginatedResponse(result))
}

// UpdatePost godoc
//
//	@Summary		Update a post
//	@Description	Update post content or visibility. Only the owner can update their post.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Post ID"
//	@Param			request	body		dto.UpdatePostRequest	true	"Update Post Request"
//	@Success		200		{object}	dto.PostResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts/{id} [patch]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var path dto.IDPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	var req dto.UpdatePostRequest
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.UpdatePostParams{
		PostID:     path.ID,
		UserID:     claims.UserID,
		Content:    req.Content,
		Visibility: (*domain.PostVisibility)(req.Visibility),
	}

	post, err := h.srv.UpdatePost(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPostResponse(post))
}

// DeletePost godoc
//
//	@Summary		Delete a post
//	@Description	Delete a post by ID. Only the owner can delete their post.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int		true	"Post ID"
//	@Success		200	{string}	string	"Deleted"
//	@Failure		400	{object}	pkg.Response
//	@Failure		401	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	claims, err := middleware.GetAuthClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path dto.IDPathParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	if err := h.srv.DeletePost(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "Deleted", nil)
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
		ID:            post.ID,
		Content:       post.Content,
		Visibility:    string(post.Visibility),
		Version:       post.Version,
		LikesCount:    post.Counts.LikesCount,
		CommentsCount: post.Counts.CommentsCount,
		SharesCount:   post.Counts.SharesCount,
		IsLiked:       post.IsLiked,
		CreatedAt:     post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     post.UpdatedAt.Format(time.RFC3339),
		User: dto.UserCompactResponse{
			ID:       post.User.ID,
			Fullname: post.User.FullName,
			Avatar:   h.urlFactory.PublicFileURL(post.User.Avatar),
		},
		Media: mediaResp,
	}
}
