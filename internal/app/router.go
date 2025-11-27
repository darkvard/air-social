package app

import (
	"context"
	"time"

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
		ok := "Ok"
		err := "Error"

		dbErr := ok
		if err := app.DB.Ping(); err != nil {
			dbErr = err.Error()
		}

		redisErr := ok
		if err := app.Redis.Ping(context.Background()).Err(); err != nil {
			redisErr = err.Error()
		}

		status := 200
		msg := ok
		if dbErr != ok || redisErr != ok {
			status = 500
			msg = err
		}

		pkg.Respond(c, pkg.Response{
			Code:    status,
			Message: msg,
			Data: gin.H{
				"status":    status,
				"db":        dbErr,
				"redis":     redisErr,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})

	})
}

func authRoutes(r *gin.Engine, h *handler.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
	}
}
