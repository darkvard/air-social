package ws

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"

	"air-social/internal/domain/chat"
)

const sendBufferSize = 256

type Hub struct {
	// a single user may have multiple connections (e.g.g multiple browser tabs)
	// guarantees each connection is a distinct key, and struct{} costs zero bytes (slice O(n) -> map O(1))
	clients    map[int64]map[*Client]struct{}
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	subMgr     *subscriptionManager
	presence   chat.PresenceStore
}

func NewHub(redis *redis.Client, presence chat.PresenceStore) *Hub {
	h := &Hub{
		clients:    make(map[int64]map[*Client]struct{}),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		broadcast:  make(chan *BroadcastMessage, 256),
		presence:   presence,
	}
	h.subMgr = newSubscriptionManager(redis)
	return h
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.add(client)
		case client := <-h.unregister:
			h.remove(client)
			client.closeChannel()
		case msg := <-h.broadcast:
			h.send(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (h *Hub) add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.userID]; !ok {
		h.clients[client.userID] = make(map[*Client]struct{})
	}
	h.clients[client.userID][client] = struct{}{}
}

func (h *Hub) remove(client *Client) {
	h.mu.Lock()
	conns, ok := h.clients[client.userID]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(conns, client)
	if len(conns) == 0 {
		delete(h.clients, client.userID)
	}
	h.mu.Unlock()

	client.mu.RLock()
	convIDs := make([]string, 0, len(client.convIDs))
	for id := range client.convIDs {
		convIDs = append(convIDs, id)
	}
	client.mu.RUnlock()

	for _, convID := range convIDs {
		if !h.hasClientInConv(convID) {
			h.subMgr.cancel(convID)
		}
	}
}

func (h *Hub) send(msg *BroadcastMessage) {
	h.mu.RLock()
	var toRemove []*Client
	for _, id := range msg.userIDs {
		for client := range h.clients[id] {
			select {
			case client.send <- msg.message:
			default:
				toRemove = append(toRemove, client)
			}
		}
	}
	h.mu.RUnlock()
	for _, client := range toRemove {
		h.remove(client)
		client.closeChannel()
	}
}

func (h *Hub) joinConv(client *Client, convID string) {
	client.joinConv(convID)
	h.subMgr.ensure(convID, func(b []byte) {
		h.broadcastToConv(convID, b)
	})
}

func (h *Hub) broadcastTyping(convID string, userID int64, typing bool) {
	eventType := EventTyping
	if !typing {
		eventType = EventStopTyping
	}
	payload, _ := OutboundEvent{
		Type: eventType,
		Data: map[string]any{keyUserID: userID},
	}.encode()
	h.broadcastToConv(convID, payload)
}

func (h *Hub) leaveConv(client *Client, convID string) {
	client.leaveConv(convID)
	if !h.hasClientInConv(convID) {
		h.subMgr.cancel(convID)
	}
}

func (h *Hub) broadcastToConv(convID string, payload []byte) {
	h.mu.RLock()
	var toRemove []*Client
	for _, conns := range h.clients {
		for client := range conns {
			client.mu.RLock()
			_, inConv := client.convIDs[convID]
			client.mu.RUnlock()
			if !inConv {
				continue
			}
			select {
			case client.send <- payload:
			default:
				toRemove = append(toRemove, client)
			}
		}
	}
	h.mu.RUnlock()
	for _, c := range toRemove {
		h.remove(c)
		c.closeChannel()
	}
}

func (h *Hub) hasClientInConv(convID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conns := range h.clients {
		for c := range conns {
			c.mu.RLock()
			_, ok := c.convIDs[convID]
			c.mu.RUnlock()
			if ok {
				return true
			}
		}
	}
	return false
}
