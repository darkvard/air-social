package route

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
)

func UserRoutes(g *gin.RouterGroup, h *handler.UserHandler, m *middleware.Manager) {
	group := g.Group(UserGroup, m.Auth)

	me := group.Group(Me)
	{
		me.GET("", h.Profile)

		json := me.Group("").Use(m.JSONOnly)
		{
			json.PATCH("", h.UpdateProfile)
			json.PUT(Password, h.ChangePassword)
			json.POST(Images, h.ConfirmFileUpload)
		}
	}
}
