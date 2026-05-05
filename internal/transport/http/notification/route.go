package notification

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, h Handler, m middleware.Manager) {
	group := g.Group("/notifications", m.Auth)
	{
		group.GET("", h.GetNotifications)
		group.GET("/unread-count", h.GetUnreadCount)
		group.PUT("/read", h.MarkAllRead)
		group.PUT("/:id/read", h.MarkOneRead)
	}
}
