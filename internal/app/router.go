package app

import "github.com/gin-gonic/gin"

func (a *Application) NewRouter() *gin.Engine {
	router := gin.Default()

	handler := a.Http.Handler

	auth := router.Group("/auth")
	{
		auth.POST("/register", handler.Auth.Register)
		auth.POST("/login", nil)
	}

	return router
}
