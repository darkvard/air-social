package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"air-social/internal/config"
	"air-social/internal/di/provider"
	"air-social/internal/transport/http/auth"
	"air-social/internal/transport/http/follow"
	"air-social/internal/transport/http/health"
	"air-social/internal/transport/http/media"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/post"
	"air-social/internal/transport/http/user"
	"air-social/pkg"
	"air-social/templates"
)

func NewServer(
	cfg config.Config,
	prov provider.Provider,
	h provider.Handler,
	mw middleware.Manager,
) *http.Server {
	router := setupEngine()
	group := router.Group(prov.Link.ApiVersion())
	{
		group.GET("", h.Health.Welcome)
		group.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		health.RegisterRoute(group, h.Health, mw)
		auth.RegisterRoute(group, h.Auth, mw)
		user.RegisterRoute(group, h.User, mw)
		media.RegisterRoute(group, h.Media, mw)
		follow.RegisterRoute(group, h.Follow, mw)
		post.RegisterRoute(group, h.Post, mw)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}
}

func setupEngine() *gin.Engine {
	e := gin.New()
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
	e.SetTrustedProxies(nil)
	e.HandleMethodNotAllowed = true

	e.SetHTMLTemplate(
		template.Must(template.New("").ParseFS(
			templates.TemplatesFS,
			"*/*.gohtml", // level 1, e.g. pages/login.gohtml
		)),
	)

	e.NoRoute(func(c *gin.Context) { pkg.NotFound(c, "Page not found") })

	e.NoMethod(func(c *gin.Context) {
		allowed := c.Writer.Header().Get("Allow")
		details := "This endpoint only supports: " + allowed
		pkg.Error(c, http.StatusMethodNotAllowed, "Method not allowed", details)
	})

	return e
}
