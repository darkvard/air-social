package ws

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

type subscriptionManager struct {
	redis *redis.Client
	mu    sync.Mutex
	subs  map[string]context.CancelFunc
}

func newSubscriptionManager(redis *redis.Client) *subscriptionManager {
	return &subscriptionManager{
		redis: redis,
		subs:  make(map[string]context.CancelFunc),
	}
}

func (m *subscriptionManager) ensure(convID string, onMessage func([]byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.subs[convID]; ok {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.subs[convID] = cancel
	go m.run(ctx, convID, onMessage)
}

func (m *subscriptionManager) cancel(convID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cancel, ok := m.subs[convID]; ok {
		cancel()
		delete(m.subs, convID)
	}
}

func (m *subscriptionManager) run(ctx context.Context, convID string, onMessage func([]byte)) {
	pubsub := m.redis.Subscribe(ctx, chatChannel(convID))
	defer pubsub.Close()
	for {
		select {
		case msg, ok := <-pubsub.Channel():
			if !ok {
				return
			}
			onMessage([]byte(msg.Payload))
		case <-ctx.Done():
			return
		}
	}
}
