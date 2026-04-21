package chat

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/transport/http/middleware"
)

func RegisterRoute(g *gin.RouterGroup, convH ConversationHandler, msgH MessageHandler, m middleware.Manager) {
	conv := g.Group("/conversations", m.Auth)
	{
		conv.POST("/direct", convH.CreateDirect)
		conv.POST("/group", convH.CreateGroup)
		conv.PATCH("/:id", convH.UpdateGroup)
	}
	{
		conv.GET("", convH.GetConversations)
		conv.GET("/:id", convH.GetConversation)
		conv.POST("/:id/accept", convH.AcceptConversation)
	}
	{
		conv.POST("/:id/members", convH.AddMember)
		conv.DELETE("/:id/members/:userID", convH.RemoveMember)
		conv.PATCH("/:id/members/:userID/role", convH.UpdateMemberRole)
	}
}
