package post

import (
	"errors"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  post.UseCase
}

func NewHandler(provider common.LinkProvider, usecase post.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
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
//	@Param			request	body		CreateRequest	true	"Create Post Request"
//	@Success		201		{object}	CreateResponse
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts [post]
func (h Handler) CreatePost(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := post.CreateParams{
		OriginalPostID: req.OriginalPostID,
		UserID:         claims.UserID,
		Content:        req.Content,
		Visibility:     post.Visibility(req.Visibility),
	}

	if len(req.Media) > 0 {
		params.Media = make([]post.MediaParams, len(req.Media))
		for i, v := range req.Media {
			params.Media[i] = h.toMediaParams(v)
		}
	}

	post, err := h.usecase.CreatePost(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			err = pkg.NewError(pkg.ErrBadRequest, "media not found")
		}
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.toCreateResponse(post))
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
//	@Success		200	{object}	PostResponse
//	@Failure		400	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id} [get]
func (h Handler) GetPost(c *gin.Context) {
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

	post, err := h.usecase.GetPostDetail(c.Request.Context(), path.ID, claims.UserID)
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
//	@Param			id		path		int		true	"User ID"
//	@Param			cursor	query		int		false	"Cursor for pagination (last post ID)"
//	@Param			limit	query		int		false	"Number of items to return"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[PostResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/users/{id}/posts [get]
func (h Handler) GetUserPosts(c *gin.Context) {
	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid user id")
		return
	}

	var req CursorQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	// todo: viewerID check isLiked
	result, err := h.usecase.GetUserPosts(c.Request.Context(), req.ToDomain(path.ID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPostListResponse(result))
}

// UpdatePost godoc
//
//	@Summary		Update a post
//	@Description	Update post content or visibility. Only the owner can update their post.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"Post ID"
//	@Param			request	body		UpdateRequest	true	"Update Post Request"
//	@Success		200		{object}	PostResponse
//	@Failure		400		{object}	pkg.Response
//	@Failure		403		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts/{id} [patch]
func (h Handler) UpdatePost(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, err.Error())
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid post id")
		return
	}

	var req UpdateRequest
	if err := pkg.StrictBindJSON(c, &req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := post.UpdateParams{
		PostID:     path.ID,
		UserID:     claims.UserID,
		Content:    req.Content,
		Visibility: (*post.Visibility)(req.Visibility),
	}

	if req.Media != nil {
		params.Media = make([]post.MediaParams, len(req.Media))
		for i, v := range req.Media {
			params.Media[i] = h.toMediaParams(v)
		}
	}

	post, err := h.usecase.UpdatePost(c.Request.Context(), params)
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
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	pkg.Response
//	@Failure		403	{object}	pkg.Response
//	@Failure		404	{object}	pkg.Response
//	@Failure		500	{object}	pkg.Response
//	@Router			/posts/{id} [delete]
func (h Handler) DeletePost(c *gin.Context) {
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

	if err := h.usecase.DeletePost(c.Request.Context(), path.ID, claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

func (h Handler) toMediaParams(req MediaItemInput) post.MediaParams {
	return post.MediaParams{
		MediaKey:  req.MediaKey,
		MediaType: req.MediaType,
		Width:     req.Width,
		Height:    req.Height,
		Duration:  req.Duration,
		Size:      req.Size,
		FileName:  req.FileName,
	}
}

func (h Handler) toCreateResponse(post *post.Post) CreateResponse {
	return CreateResponse{
		ID:             post.ID,
		OriginalPostID: post.OriginalPostID,
		Content:        post.Content,
		Visibility:     string(post.Visibility),
		CreatedAt:      post.CreatedAt,
		Media:          h.toMediaItemResponse(post.Media),
	}
}

func (h Handler) toMediaItemResponse(media []post.Media) []MediaItemResponse {
	result := make([]MediaItemResponse, len(media))
	if media == nil {
		return result
	}
	for i, m := range media {
		result[i] = MediaItemResponse{
			ID:        m.ID,
			URL:       h.provider.PublicFile(m.MediaKey),
			MediaType: m.MediaType,
			Width:     m.Metadata.Width,
			Height:    m.Metadata.Height,
			Duration:  m.Metadata.Duration,
			FileName:  m.Metadata.FileName,
		}
	}
	return result
}

func (h Handler) toPostResponse(post *post.Post) PostResponse {
	if post == nil {
		return PostResponse{}
	}

	var author *UserResponse
	if post.Author != nil {
		author = &UserResponse{
			ID:         post.Author.ID,
			Fullname:   post.Author.FullName,
			Avatar:     h.provider.PublicFile(post.Author.Avatar),
			IsVerified: post.Author.IsVerified,
		}
	}

	var viewerLiked *bool
	if post.IsLiked != nil {
		viewerLiked = post.IsLiked
	}

	return PostResponse{
		ID:             post.ID,
		OriginalPostID: post.OriginalPostID,
		Content:        post.Content,
		Visibility:     string(post.Visibility),
		LikesCount:     post.Stat.LikesCount,
		CommentsCount:  post.Stat.CommentsCount,
		SharesCount:    post.Stat.SharesCount,
		CreatedAt:      post.CreatedAt,
		UpdatedAt:      post.UpdatedAt,
		Media:          h.toMediaItemResponse(post.Media),
		User:           author,
		IsLiked:        viewerLiked,
	}
}

func (h Handler) toPostListResponse(result common.CursorPaginatedResult[post.Post, int64]) CursorPaginatedResponse[PostResponse] {
	data := make([]PostResponse, len(result.Data))
	for i := range result.Data {
		data[i] = h.toPostResponse(&result.Data[i])
	}

	return CursorPaginatedResponse[PostResponse]{
		Data: data,
		Meta: MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}

// GetPostSharers godoc
//
//	@Summary		Get users who shared a post
//	@Description	Get a list of users who shared a specific post using cursor-based pagination.
//	@Tags			Post
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Post ID"
//	@Param			cursor	query		int		false	"Cursor for pagination (last share post ID)"
//	@Param			limit	query		int		false	"Number of items to return"
//	@Param			sort	query		string	false	"Sort order"	Enums(latest, oldest)
//	@Success		200		{object}	CursorPaginatedResponse[SharerResponse]
//	@Failure		400		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/posts/{id}/shares [get]
func (h Handler) GetPostSharers(c *gin.Context) {
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

	result, err := h.usecase.GetPostSharers(c.Request.Context(), req.ToGetSharersParams(path.ID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toSharerListResponse(result))
}

func (h Handler) toSharerResponse(s post.Sharer) SharerResponse {
	return SharerResponse{
		ID:         s.UserID,
		FullName:   s.FullName,
		Avatar:     h.provider.PublicFile(s.Avatar),
		IsVerified: s.IsVerified,
	}
}

func (h Handler) toSharerListResponse(result common.CursorPaginatedResult[post.Sharer, int64]) CursorPaginatedResponse[SharerResponse] {
	data := make([]SharerResponse, len(result.Data))
	for i, v := range result.Data {
		data[i] = h.toSharerResponse(v)
	}

	return CursorPaginatedResponse[SharerResponse]{
		Data: data,
		Meta: MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}
