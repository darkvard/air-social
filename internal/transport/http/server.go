package http

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
	"air-social/templates"
)

const (
	Health     = "/health"
	SwaggerAny = "/swagger/*any"
)

const (
	AuthGroup      = "/auth"
	Register       = "/register"
	Login          = "/login"
	Refresh        = "/refresh"
	ResetPassword  = "/reset-password"
	ForgotPassword = "/forgot-password"
	VerifyEmail    = "/verify-email"
	Logout         = "/logout"
)

const (
	UserGroup  = "/users"
	Me         = "/me"
	Password   = "/password"
	Followers  = "/:id/followers"
	Followings = "/:id/followings"
	FollowUser = "/:id/follow"
)

const (
	MediaGroup      = "/media"
	PresignedUpload = "/presigned"
	Images          = "/images"
)

func NewServer(
	cfg config.Config,
	urls domain.URLFactory,
	mw *middleware.Manager,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	mediaH *handler.MediaHandler,
	healthH *handler.HealthHandler,
	followH *handler.FollowHandler,
) *http.Server {
	e := setupEngine()

	v := e.Group(urls.APIRouterPath())
	{
		commonRoutes(v, healthH, mw)
		authRoutes(v, authH, mw)
		userRoutes(v, userH, mw)
		mediaRoutes(v, mediaH, mw)
		followRoutes(v, followH, mw)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      e,
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

func commonRoutes(g *gin.RouterGroup, h *handler.HealthHandler, m *middleware.Manager) {
	g.GET("", h.Welcome)
	g.GET(Health, m.Basic, h.HealthCheck)
	g.GET(SwaggerAny, ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func authRoutes(g *gin.RouterGroup, h *handler.AuthHandler, m *middleware.Manager) {
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

func userRoutes(g *gin.RouterGroup, h *handler.UserHandler, m *middleware.Manager) {
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

func mediaRoutes(g *gin.RouterGroup, h *handler.MediaHandler, m *middleware.Manager) {
	auth := g.Group(MediaGroup, m.Auth)
	auth.POST(PresignedUpload, h.PresignedUpload)
}

func followRoutes(g *gin.RouterGroup, h *handler.FollowHandler, m *middleware.Manager) {
	group := g.Group(UserGroup)
	{
		group.GET(Followers, h.GetFollowers)
		group.GET(Followings, h.GetFollowings)
	}

	auth := group.Group("", m.Auth)
	{
		auth.POST(FollowUser, h.Follow)
		auth.DELETE(FollowUser, h.Unfollow)
	}
}
