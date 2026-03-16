package like

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("", m.Auth)
	postGroup := group.Group("/posts")
	{
		postGroup.GET("/:id/likes", h.GetPostLikers)
		postGroup.POST("/:id/likes", h.LikePost)
		postGroup.DELETE("/:id/likes", h.UnlikePost)
	}
	commentGroup := group.Group("/comments")
	{
		commentGroup.GET("/:id/likes", h.GetCommentLikers)
		commentGroup.POST("/:id/likes", h.LikeComment)
		commentGroup.DELETE("/:id/likes", h.UnlikeComment)
	}
}
