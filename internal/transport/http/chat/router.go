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
	{
		conv.GET("/:id/messages", msgH.GetMessages)
		conv.POST("/:id/messages", msgH.SendMessage)
		conv.PATCH("/:id/messages/:msgID", msgH.EditMessage)
		conv.DELETE("/:id/messages/:msgID", msgH.DeleteMessage)
	}
	{
		conv.PATCH("/:id/read", msgH.MarkRead)
		conv.DELETE("/:id/read", msgH.MarkUnread)
	}
	{
		conv.POST("/:id/messages/:msgID/reactions", msgH.AddReaction)
		conv.DELETE("/:id/messages/:msgID/reactions", msgH.RemoveReaction)
	}
}
