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
	"air-social/internal/di"
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


func NewServer(cfg config.Config, svc *di.ServiceContainer, ifc *di.InfraContainer) *http.Server {
	e := setupEngine()
	mw := middleware.NewManager(cfg.Server, svc.Token)
	v := e.Group("/" + cfg.Server.Version)
	{
		commonRoutes(v, ifc, mw)
		authRoutes(v, svc, mw)
		userRoutes(v, svc, mw)
		mediaRoutes(v, svc, mw)
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

func commonRoutes(rg *gin.RouterGroup, ifc *di.InfraContainer, mw *middleware.Manager) {
	h := handler.NewHealthHandler(ifc.DB, ifc.Redis, ifc.Rabbit, ifc.Minio)
	r := rg.Group("")
	{
		r.GET(Health, mw.Basic, h.HealthCheck)
		r.GET(SwaggerAny, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func authRoutes(rg *gin.RouterGroup, svc *di.ServiceContainer, mw *middleware.Manager) {
	h := handler.NewAuthHandler(svc.Auth)
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

func userRoutes(rg *gin.RouterGroup, svc *di.ServiceContainer, mw *middleware.Manager) {
	h := handler.NewUserHandler(svc.User)
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

func mediaRoutes(rg *gin.RouterGroup, svc *di.ServiceContainer, mw *middleware.Manager) {
	h := handler.NewMediaHandler(svc.Media)
	m := rg.Group(MediaGroup, mw.Auth)
	{
		m.POST(PresignedUpload, h.PresignedUpload)
	}
}
