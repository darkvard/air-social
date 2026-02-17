package user

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h *handler.UserHandler, m *middleware.Manager) {
	group := g.Group("/users", m.Auth)
	me := group.Group("/me")
	{
		me.GET("", h.Profile)

		json := me.Group("").Use(m.JSONOnly)
		{
			json.PATCH("", h.UpdateProfile)
			json.PUT("/password", h.ChangePassword)
			json.POST("/images", h.ConfirmFileUpload)
		}
	}
}
