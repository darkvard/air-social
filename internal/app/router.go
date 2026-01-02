package app

import (
	"html/template"

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

	basic := gin.BasicAuth(
		gin.Accounts{
			a.Config.Server.Username: a.Config.Server.Password,
		},
	)
	auth := middleware.Auth(s.Token)

	v1 := e.Group("/" + a.Config.Server.Version)
	{
		commonRoutes(v1, h.Health, basic)
		authRoutes(v1, h.Auth, auth)
		userRoutes(v1, h.User, auth)
	}
	return e
}

func (a *Application) setupEngine() *gin.Engine {
	e := gin.New()
	e.Use(gin.Logger())
	e.Use(gin.Recovery())
	e.SetTrustedProxies(nil)
	e.SetHTMLTemplate(
		template.Must(template.New("").ParseFS(
			templates.TemplatesFS,
			"*/*.gohtml", // level 1, e.g. pages/login.gohtml
		)),
	)
	e.NoRoute(func(c *gin.Context) { pkg.NotFound(c, "Page not found") })
	return e
}

func commonRoutes(rg *gin.RouterGroup, h *handler.HealthHandler, basic gin.HandlerFunc) {
	r := rg.Group("")
	{
		r.GET(routes.Health, basic, h.HealthCheck)
		r.GET(routes.SwaggerAny, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func authRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, auth gin.HandlerFunc) {
	publish := rg.Group(routes.AuthGroup)
	{
		publish.POST(routes.Register, h.Register)
		publish.POST(routes.Login, h.Login)
		publish.POST(routes.Refresh, h.Refresh)
		publish.POST(routes.ForgotPassword, h.ForgotPassword)
		publish.GET(routes.ResetPassword, h.ShowResetPasswordPage)
		publish.POST(routes.ResetPassword, h.ResetPassword)
		publish.GET(routes.VerifyEmail, h.VerifyEmail)
	}
	protected := publish.Group("").Use(auth)
	{
		protected.POST(routes.Logout, h.Logout)
	}
}

func userRoutes(rg *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc) {
	protected := rg.Group(routes.UserGroup, auth)
	{
		protected.GET(routes.Me, h.Profile)
		protected.PUT(routes.Me, h.UpdateProfile)
		protected.PUT(routes.Password, h.ChangePassword)
		protected.POST(routes.Avatar, h.UpdateAvatar)
	}
}
