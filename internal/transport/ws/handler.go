package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // dev only
	},
}

func ServeWS(hub *Hub, c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		pkg.Log().Errorw("ws upgrade error", "userID", claims.UserID, "error", err)
		return
	}

	client := &Client{
		userID: claims.UserID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
	}
	pkg.Log().Infow("ws client connected", "userID", client.userID)

	hub.register <- client

	go client.readPump()
	go client.writePump()
}
