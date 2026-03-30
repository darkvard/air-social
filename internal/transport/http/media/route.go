package media

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	auth := g.Group("/media", m.Auth)
	auth.POST("/presigned-urls", m.RateLimitMediaPresigned(), h.PresignedUpload)
}
