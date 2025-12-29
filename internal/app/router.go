package app

import (
	"html/template"

	"github.com/gin-gonic/gin"

	"air-social/internal/routes"
	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
	"air-social/templates"
)

func (a *Application) NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.SetTrustedProxies(nil)
	r.SetHTMLTemplate(
		template.Must(template.New("").ParseFS(
			templates.TemplatesFS,
			"*/*.gohtml",     // level 1, e.g. pages/login.gohtml
		)),
	)


	h := a.Http.Handler
	s := a.Http.Service
	authMiddleware := middleware.AuthMiddleware(s.Token)

	a.commonRoutes(r)

	v1 := r.Group("/" + a.Config.Server.Version)
	{
		authRoutes(v1, h.Auth, authMiddleware)
		// userRoutes(v1, h.User, authMiddleware())
	}

	return r
}

func (app *Application) commonRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		pkg.NotFound(c, "Page not found")
	})

	r.GET(routes.Health, gin.BasicAuth(
		gin.Accounts{app.Config.Server.Username: app.Config.Server.Password},
	), func(c *gin.Context) {
		pkg.Success(c, app.HealthStatus())
	})
}

func authRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group(routes.AuthGroup)
	{
		auth.POST(routes.Register, h.Register)
		auth.POST(routes.Login, h.Login)
		auth.POST(routes.Refresh, h.Refresh)
		auth.POST(routes.ResetPassword, h.ResetPassword)
		auth.POST(routes.ForgotPassword, h.ForgotPassword)
		auth.GET(routes.VerifyEmail, h.VerifyEmail)
	}
	protected := auth.Group("").Use(authMiddleware)
	{
		protected.POST(routes.Logout, h.Logout)
	}
}

// func userRoutes(rg *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc) {
// 	users := rg.Group("/users", auth)
// 	{
// 		me := users.Group("/me")
// 		{
// 			me.GET("", h.GetProfile)
// 			me.PUT("", h.UpdateProfile)
// 			me.PUT("/password", h.ChangePassword)
// 			me.POST("/avatar", h.UpdateAvatar)
// 		}
// 	}
// }
