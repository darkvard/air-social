package feed

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	feed := g.Group("/feed", m.Auth)
	{
		feed.GET("", h.GetNewsfeed)
	}
}
