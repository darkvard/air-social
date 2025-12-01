package app

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/pkg"
)

func (a *Application) NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.SetTrustedProxies(nil)

	h := a.Http.Handler
	a.commonRoutes(r)
	authRoutes(r, h.Auth)

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

func authRoutes(r *gin.Engine, h *handler.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", nil)
		auth.POST("/logout", nil)
	}
}
