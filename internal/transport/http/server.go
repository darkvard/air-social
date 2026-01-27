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
	UserGroup    = "/users"
	Me           = "/me"
	Password     = "/password"
	ProfileImage = "/profile-image"
)

const (
	MediaGroup      = "/media"
	PresignedUpload = "/presigned"
	ConfirmUpload   = "/confirm"
)

func NewServer(
	cfg config.Config,
	urls domain.URLFactory,
	mw *middleware.Manager,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	mediaH *handler.MediaHandler,
	healthH *handler.HealthHandler,
) *http.Server {
	e := setupEngine()

	v := e.Group(urls.APIRouterPath())
	{
		commonRoutes(v, healthH, mw)
		authRoutes(v, authH, mw)
		userRoutes(v, userH, mw)
		mediaRoutes(v, mediaH, mw)
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      e,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
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
		allowedMethods := c.Writer.Header().Get("Allow")
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code":    http.StatusMethodNotAllowed,
			"message": "Method not allowed",
			"details": "This endpoint only supports: " + allowedMethods,
		})
	})

	return e
}

func commonRoutes(rg *gin.RouterGroup, h *handler.HealthHandler, mw *middleware.Manager) {
	{
		rg.GET("", h.Welcome)
		rg.GET(Health, mw.Basic, h.HealthCheck)
		rg.GET(SwaggerAny, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func authRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, mw *middleware.Manager) {
	a := rg.Group(AuthGroup)
	{
		a.GET(ResetPassword, h.ShowResetPasswordPage)
		a.GET(VerifyEmail, h.VerifyEmail)

		j := a.Group("").Use(mw.JSONOnly)
		{
			j.POST(Register, h.Register)
			j.POST(Login, h.Login)
			j.POST(Refresh, h.Refresh)
			j.POST(ForgotPassword, h.ForgotPassword)
			j.POST(ResetPassword, h.ResetPassword)
		}
		p := a.Group("").Use(mw.Auth)
		{
			p.POST(Logout, h.Logout)
		}
	}
}

func userRoutes(rg *gin.RouterGroup, h *handler.UserHandler, mw *middleware.Manager) {
	p := rg.Group(UserGroup, mw.Auth)
	{
		p.GET(Me, h.Profile)

		j := p.Group("").Use(mw.JSONOnly)
		{
			j.PUT(Password, h.ChangePassword)
			j.PATCH(Me, h.UpdateProfile)
			j.POST(ProfileImage+ConfirmUpload, h.ConfirmFileUpload)
		}
	}
}

func mediaRoutes(rg *gin.RouterGroup, h *handler.MediaHandler, mw *middleware.Manager) {
	m := rg.Group(MediaGroup, mw.Auth)
	{
		m.POST(PresignedUpload, h.PresignedUpload)
	}
}
