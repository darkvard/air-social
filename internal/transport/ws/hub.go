package ws

import (
	"context"
	"sync"
)

const sendBufferSize = 256

type Hub struct {
	// a single user may have multiple connections (e.g.g multiple browser tabs)
	// guarantees each connection is a distinct key, and struct{} costs zero bytes (slice O(n) -> map O(1))
	clients map[int64]map[*Client]struct{}
	mu      sync.RWMutex

	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]struct{}),
		register:   make(chan *Client, 16),
		unregister: make(chan *Client, 16),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.add(client)
		case client := <-h.unregister:
			h.remove(client)
			close(client.send)
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
	defer h.mu.Unlock()
	conns, ok := h.clients[client.userID]
	if !ok {
		return
	}
	delete(conns, client)
	if len(conns) == 0 {
		delete(h.clients, client.userID)
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
		close(client.send)
	}
}
