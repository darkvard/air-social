package route

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
)

func AuthRoutes(g *gin.RouterGroup, h *handler.AuthHandler, m *middleware.Manager) {
	group := g.Group(AuthGroup)
	{
		group.GET(ResetPassword, h.ShowResetPasswordPage)
		group.GET(VerifyEmail, h.VerifyEmail)
	}

	json := group.Group("").Use(m.JSONOnly)
	{
		json.POST(Register, h.Register)
		json.POST(Login, h.Login)
		json.POST(Refresh, h.Refresh)
		json.POST(ForgotPassword, h.ForgotPassword)
		json.POST(ResetPassword, h.ResetPassword)
	}

	auth := group.Group("").Use(m.Auth)
	auth.POST(Logout, h.Logout)
}
