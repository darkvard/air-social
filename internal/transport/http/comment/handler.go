package comment

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/comment"
	"air-social/internal/domain/common"
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

func (h Handler) CreateComment(c *gin.Context) {

}

func (h Handler) GetComments(c *gin.Context) {

}

func (h Handler) UpdateComment(c *gin.Context) {

}

func (h Handler) DeleteComment(c *gin.Context) {

}

func (h Handler) GetReplies(c *gin.Context) {

}