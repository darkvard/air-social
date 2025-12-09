package app

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

func (a *Application) NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	h := a.Http.Handler
	s := a.Http.Service
	authMiddleware := middleware.AuthMiddleware(s.Token)

	a.commonRoutes(r)

	v1 := r.Group("/v1")
	{
		authRoutes(v1, h.Auth, authMiddleware)
		// userRoutes(v1, h.User, authMiddleware())
	}

	return r
}

func (app *Application) commonRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		pkg.NotFound(c, "Page not found")
	})

	r.GET("/health", func(c *gin.Context) {
		pkg.Success(c, app.HealthStatus())
	})
}

func authRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/reset-password", h.ResetPassword)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.GET("/verify-email", h.VerifyEmail)
	}
	protected := auth.Group("").Use(authMiddleware)
	{
		protected.POST("/logout", h.Logout)
	}
}

// func userRoutes(rg *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc) {
// 	users := rg.Group("/users", auth)
// 	{
// 		me := users.Group("/me")
// 		{
// 			me.GET("", h.GetProfile)
// 			me.PUT("", h.UpdateProfile)
// 			me.PUT("/password", h.ChangePassword)
// 			me.POST("/avatar", h.UpdateAvatar)
// 		}
// 	}
// }
