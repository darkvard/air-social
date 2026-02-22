package user

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("/users/me", m.Auth)
	{
		group.GET("", h.Profile)

		json := group.Group("").Use(m.JSONOnly)
		{
			json.PATCH("", h.UpdateProfile)
			json.PUT("/password", h.ChangePassword)
			json.POST("/avatar", h.UpdateAvatar)
			json.POST("/image-cover", h.UpdateImageCover)
		}
	}
}
