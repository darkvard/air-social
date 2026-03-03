package comment

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("", m.Auth)
	postGroup := group.Group("/posts")
	{
		postGroup.POST("/:id/comments", h.CreateComment)
		postGroup.GET("/:id/comments", h.GetComments)
	}
	commentGroup := group.Group("/comments")
	{
		commentGroup.GET("/:id/replies", h.GetReplies)
		commentGroup.PATCH("/:id", h.UpdateComment)
		commentGroup.DELETE("/:id", h.DeleteComment)
	}
}
