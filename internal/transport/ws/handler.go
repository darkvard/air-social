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

// ServeWS godoc
//
//	@Summary		Connect to real-time chat
//	@Description	Upgrades the HTTP connection to WebSocket. Requires a valid Bearer token.
//	@Description
//	@Description	**Authentication**: Pass token via header `Authorization: Bearer <token>`.
//	@Description
//	@Description	## Inbound events (client → server)
//	@Description
//	@Description	| type | payload | description |
//	@Description	|------|---------|-------------|
//	@Description	| `ping` | — | Heartbeat. Server replies with `pong` and marks user online. |
//	@Description	| `join` | `{"conversation_id":"..."}` | Subscribe to real-time events for a conversation. Must call before receiving messages. |
//	@Description	| `leave` | `{"conversation_id":"..."}` | Unsubscribe from a conversation. |
//	@Description	| `typing` | `{"conversation_id":"..."}` | Notify others that user is typing. |
//	@Description	| `stop_typing` | `{"conversation_id":"..."}` | Notify others that user stopped typing. |
//	@Description
//	@Description	## Outbound events (server → client)
//	@Description
//	@Description	| type | data | description |
//	@Description	|------|------|-------------|
//	@Description	| `pong` | — | Response to `ping`. |
//	@Description	| `new_message` | `Message` | A new message was sent in a joined conversation. |
//	@Description	| `message_edited` | `Message` | A message was edited. |
//	@Description	| `message_deleted` | `{"message_id":"..."}` | A message was soft-deleted. |
//	@Description	| `reaction_added` | `{"message_id":"...","reaction":{...}}` | Reaction added to a message. |
//	@Description	| `reaction_removed` | `{"message_id":"...","user_id":...}` | Reaction removed. |
//	@Description	| `read` | `{"user_id":...,"last_read_msg_id":"..."}` | A participant marked the conversation as read. |
//	@Description	| `typing` | `{"user_id":...}` | A participant is typing. |
//	@Description	| `stop_typing` | `{"user_id":...}` | A participant stopped typing. |
//	@Description
//	@Description	## Typical flow
//	@Description	1. Connect with valid Bearer token
//	@Description	2. Send `{"type":"join","payload":{"conversation_id":"<id>"}}` to subscribe
//	@Description	3. Use REST `POST /conversations/{id}/messages` to send a message
//	@Description	4. Receive `new_message`, `typing`, etc. events in real time
//	@Tags			websocket
//	@Security		BearerAuth
//	@Success		101	"Switching Protocols"
//	@Failure		401	{object}	pkg.Response
//	@Router			/ws [get]
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

	client := newClient(claims.UserID, hub, conn)
	pkg.Log().Infow("ws client connected", "userID", client.userID)

	hub.register <- client

	go client.readPump()
	go client.writePump()
}
