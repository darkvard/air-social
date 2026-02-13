package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	"air-social/internal/di/provider"
	"air-social/internal/domain"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/route"
	"air-social/pkg"
	"air-social/templates"
)

func NewServer(
	cfg config.Config,
	urls domain.URLFactory,
	mw *middleware.Manager,
	handler *provider.Handlers,
) *http.Server {
	engine := setupEngine()

	group := engine.Group(urls.ApiPath())
	{
		route.CommonRoutes(group, handler.Health, mw)
		route.AuthRoutes(group, handler.Auth, mw)
		route.UserRoutes(group, handler.User, mw)
		route.MediaRoutes(group, handler.Media, mw)
		route.FollowRoutes(group, handler.Follow, mw)
		route.PostRoutes(group, handler.Post, mw)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      engine,
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
