package media

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("/media", m.Auth)
	group.GET("/presigned-urls", m.Basic, h.PresignedUpload)
}
