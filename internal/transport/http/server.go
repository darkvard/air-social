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

func NewServer(cfg config.Config, svc *di.ServiceContainer, ifc *di.InfraContainer) *http.Server {
	e := setupEngine()
	mw := middleware.NewManager(cfg.Server, svc.Token)
	v := e.Group("/" + cfg.Server.Version)
	{
		commonRoutes(v, ifc, mw)
		authRoutes(v, svc, mw)
		userRoutes(v, svc, mw)
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
	auth := rg.Group(AuthGroup)
	{
		auth.GET(ResetPassword, h.ShowResetPasswordPage)
		auth.GET(VerifyEmail, h.VerifyEmail)

		jmw := auth.Group("").Use(mw.JSONOnly)
		{
			jmw.POST(Register, h.Register)
			jmw.POST(Login, h.Login)
			jmw.POST(Refresh, h.Refresh)
			jmw.POST(ForgotPassword, h.ForgotPassword)
			jmw.POST(ResetPassword, h.ResetPassword)
		}
		protected := auth.Group("").Use(mw.Auth)
		{
			protected.POST(Logout, h.Logout)
		}
	}
}

func userRoutes(rg *gin.RouterGroup, svc *di.ServiceContainer, mw *middleware.Manager) {
	h := handler.NewUserHandler(svc.User)
	protected := rg.Group(UserGroup, mw.Auth)
	{
		protected.GET(Me, h.Profile)

		jmw := protected.Group("").Use(mw.JSONOnly)
		{
			jmw.PUT(Password, h.ChangePassword)
			jmw.PATCH(Me, h.UpdateProfile)
		}
		fmw := protected.Group("").Use(mw.MultipartOnly)
		{
			fmw.POST(Avatar, h.UpdateAvatar)
		}
	}
}
