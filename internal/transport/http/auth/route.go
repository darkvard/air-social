package auth

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h *handler.AuthHandler, m *middleware.Manager) {
	group := g.Group("/auth")
	{
		group.GET("/reset-password", h.ShowResetPasswordPage)
		group.GET("/verify-email", h.VerifyEmail)
	}

	json := group.Group("").Use(m.JSONOnly)
	{
		json.POST("/register", h.Register)
		json.POST("/login", h.Login)
		json.POST("/refresh-token", h.Refresh)
		json.POST("/forgot-password", h.ForgotPassword)
		json.POST("/reset-password", h.ResetPassword)
	}

	auth := group.Group("").Use(m.Auth)
	auth.POST("/logout", h.Logout)
}
