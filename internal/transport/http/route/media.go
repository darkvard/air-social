package route

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/handler"
	"air-social/internal/transport/http/middleware"
)

func MediaRoutes(g *gin.RouterGroup, h *handler.MediaHandler, m *middleware.Manager) {
	auth := g.Group(MediaGroup, m.Auth)
	auth.POST(PresignedUpload, h.PresignedUpload)
}
