package post

import (
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
		UserID:     claims.UserID,
		Content:    req.Content,
		Visibility: post.Visibility(req.Visibility),
	}

	if len(req.Media) > 0 {
		params.Media = make([]post.MediaParams, len(req.Media))
		for i, v := range req.Media {
			params.Media[i] = h.toMediaParams(v)
		}
	}

	post, err := h.usecase.CreatePost(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.toCreateResponse(post))
}

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

	_ = common.NewCursorPaginatedResult(result.Data, req.Limit)
	// todo: map to dto
}

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

	post, err := h.usecase.UpdatePost(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toPostResponse(post))
}

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
		ID:         post.ID,
		Content:    post.Content,
		Visibility: string(post.Visibility),
		CreatedAt:  post.CreatedAt,
		Media:      h.toMediaItemResponse(post.Media),
	}
}

func (h Handler) toMediaItemResponse(media []post.Media) []MediaItemResponse {
	mediaResp := make([]MediaItemResponse, len(media))
	for i, m := range media {
		mediaResp[i] = MediaItemResponse{
			ID:        m.ID,
			URL:       h.provider.PublicFile(m.MediaKey),
			MediaType: m.MediaType,
			Width:     m.Metadata.Width,
			Height:    m.Metadata.Height,
			Duration:  m.Metadata.Duration,
			FileName:  m.Metadata.FileName,
		}
	}
	return mediaResp
}

func (h Handler) toPostResponse(post *post.Post) PostResponse {
	return PostResponse{
		ID:            post.ID,
		Content:       post.Content,
		Visibility:    string(post.Visibility),
		LikesCount:    post.Stat.LikesCount,
		CommentsCount: post.Stat.CommentsCount,
		SharesCount:   post.Stat.SharesCount,
		IsLiked:       post.IsLiked,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		User: UserResponse{
			ID:       post.Author.ID,
			Fullname: post.Author.FullName,
			Avatar:   h.provider.PublicFile(post.Author.Avatar),
		},
		Media: h.toMediaItemResponse(post.Media),
	}
}
