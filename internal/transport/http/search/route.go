package search

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("/search", m.Auth)
	group.GET("/users", h.SearchUsers)
	group.GET("/posts", h.SearchPosts)
}
