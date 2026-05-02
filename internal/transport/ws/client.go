package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
	maxMsgSize = 4096
)

type Client struct {
	userID    int64
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	convIDs   map[string]struct{}
	mu        sync.RWMutex
	closeOnce sync.Once
}


func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()

	c.conn.SetReadLimit(maxMsgSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var event InboundEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue
		}

		switch event.Type {
		case EventPing:
			pong, _ := OutboundEvent{Type: EventPong}.encode()
			c.send <- pong
			go func() { _ = c.hub.presence.SetOnline(context.Background(), c.userID) }()

		case EventJoin:
			if id := c.convIDFromPayload(event.Payload); id != "" {
				c.hub.joinConv(c, id)
			}
		case EventLeave:
			if id := c.convIDFromPayload(event.Payload); id != "" {
				c.hub.leaveConv(c, id)
			}
		case EventTyping:
			if id := c.convIDFromPayload(event.Payload); id != "" {
				c.hub.broadcastTyping(id, c.userID, true)
			}
		case EventStopTyping:
			if id := c.convIDFromPayload(event.Payload); id != "" {
				c.hub.broadcastTyping(id, c.userID, false)
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) closeChannel() {
	c.closeOnce.Do(func() { close(c.send) })
}

func (c *Client) convIDFromPayload(raw json.RawMessage) string {
	var p convPayload
	json.Unmarshal(raw, &p)
	return p.ConversationID
}

func (c *Client) joinConv(convID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.convIDs[convID] = struct{}{}
}

func (c *Client) leaveConv(convID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.convIDs, convID)
}
