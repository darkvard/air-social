package post

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	post := g.Group("/posts", m.Auth)
	{
		post.GET("/:id", h.GetPost)
		post.DELETE("/:id", h.DeletePost)

		json := post.Group("").Use(m.JSONOnly)
		{
			json.PATCH("/:id", h.UpdatePost)
			json.POST("", h.CreatePost)
		}
	}

	user := g.Group("/users", m.Auth)
	{
		user.GET("/:id/posts", h.GetUserPosts)
	}
}
