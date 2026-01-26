package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
	"air-social/pkg"
)

func NewConnection(cfg config.RabbitMQConfig) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = pkg.Retry(ctx, 10, 2*time.Second, func() error {
		conn, err = amqp.Dial(cfg.URL)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("rabbitmq: %w", err)
	}
	return conn, nil
}

func NewEventPublisher(conn *amqp.Connection) (*Publisher, error) {
	if conn == nil {
		return nil, errors.New("rabbitmq connection cannot nil")
	}

	pub, err := NewPublisher(
		conn,
		EventsExchange,
		10,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq init publisher failed: %w", err)
	}
	return pub, nil
}

type HealthChecker struct {
	Conn *amqp.Connection
	URL  string
	mu   sync.Mutex
}

func (h *HealthChecker) Ping() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.Conn == nil || h.Conn.IsClosed() {
		conn, err := amqp.Dial(h.URL)
		if err != nil {
			return fmt.Errorf("connection closed and reconnect failed: %w", err)
		}
		h.Conn = conn
	}
	return nil
}
