package follow

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("/users").Group("", m.Auth)
	{
		group.POST("/:id/follow", h.Follow)
		group.DELETE("/:id/follow", h.Unfollow)
		group.GET("/:id/followers", h.GetFollowers)
		group.GET("/:id/followings", h.GetFollowings)
	}
}
