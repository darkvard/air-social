package rabbitmq

import (
	"context"
	"errors"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
)

type Health struct {
	Conn *amqp.Connection
	URL  string
	mu   sync.Mutex
}

func NewHealth(conn *amqp.Connection, cfg config.RabbitMQConfig) *Health {
	return &Health{
		Conn: conn,
		URL:  cfg.URL,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.Conn == nil || h.Conn.IsClosed() {
		return errors.New("rabbit connection closed")
	}
	return nil
}
