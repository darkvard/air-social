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
	"air-social/internal/transport/http/chat"
	"air-social/internal/transport/http/comment"
	"air-social/internal/transport/http/feed"
	"air-social/internal/transport/http/follow"
	"air-social/internal/transport/http/health"
	"air-social/internal/transport/http/like"
	"air-social/internal/transport/http/media"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/post"
	"air-social/internal/transport/http/search"
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
	pkg.SetupValidator()
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
		like.RegisterRoute(group, h.Like, mw)
		comment.RegisterRoute(group, h.Comment, mw)
		feed.RegisterRoute(group, h.Feed, mw)
		search.RegisterRoute(group, h.Search, mw)
		chat.RegisterRoute(group, h.Conversation, h.Message, mw)
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
	e.Use(middleware.Recovery())
	e.SetTrustedProxies([]string{"172.16.0.0/12", "127.0.0.1"}) // docker bridge subnet + localhost
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
