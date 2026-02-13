package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
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

func (h *PostHandler) CreatePost(c *gin.Context) {

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
            Username: post.User.Username,
            Avatar:   h.urlFactory.PublicFileURL(post.User.Profile.Avatar),  
        },
        Media: mediaResp,
    }
}