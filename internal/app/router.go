package app

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"air-social/internal/routes"
	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
	"air-social/templates"
)

func (a *Application) NewRouter() *gin.Engine {
	e := a.setupEngine()
	h := a.Http.Handler
	s := a.Http.Service

	mw := middleware.NewManager(a.Config.Server, s.Token)

	v1 := e.Group("/" + a.Config.Server.Version)
	{
		commonRoutes(v1, h.Health, mw)
		authRoutes(v1, h.Auth, mw)
		userRoutes(v1, h.User, mw)
	}
	return e
}

func (a *Application) setupEngine() *gin.Engine {
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
	r := rg.Group("")
	{
		r.GET(routes.Health, mw.Basic, h.HealthCheck)
		r.GET(routes.SwaggerAny, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func authRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, mw *middleware.Manager) {
	auth := rg.Group(routes.AuthGroup)
	{
		auth.GET(routes.ResetPassword, h.ShowResetPasswordPage)
		auth.GET(routes.VerifyEmail, h.VerifyEmail)

		jwm := auth.Group("").Use(mw.JSONOnly)
		{
			jwm.POST(routes.Register, h.Register)
			jwm.POST(routes.Login, h.Login)
			jwm.POST(routes.Refresh, h.Refresh)
			jwm.POST(routes.ForgotPassword, h.ForgotPassword)
			jwm.POST(routes.ResetPassword, h.ResetPassword)
		}
		protected := auth.Group("").Use(mw.Auth)
		{
			protected.POST(routes.Logout, h.Logout)
		}
	}
}

func userRoutes(rg *gin.RouterGroup, h *handler.UserHandler, mw *middleware.Manager) {
	protected := rg.Group(routes.UserGroup, mw.Auth)
	{
		protected.GET(routes.Me, h.Profile)

		jmw := protected.Group("").Use(mw.JSONOnly)
		{
			jmw.PUT(routes.Password, h.ChangePassword)
			jmw.PATCH(routes.Me, h.UpdateProfile)
		}
		fmw := protected.Group("").Use(mw.MultipartOnly)
		{
			fmw.POST(routes.Avatar, h.UpdateAvatar)
		}
	}
}
