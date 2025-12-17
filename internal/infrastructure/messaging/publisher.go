package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type pubChannel struct {
	ch       *amqp.Channel
	confirms chan amqp.Confirmation
}

type Publisher struct {
	conn   *amqp.Connection
	cfg    ExchangeConfig
	chPool chan *pubChannel
	once   sync.Once
}

// NewPublisher creates a publisher with channel pool + confirm mode
func NewPublisher(
	conn *amqp.Connection,
	exCfg ExchangeConfig,
	poolSize int,
) (*Publisher, error) {

	if poolSize <= 0 {
		poolSize = 1
	}

	p := &Publisher{
		conn:   conn,
		cfg:    exCfg,
		chPool: make(chan *pubChannel, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		ch, err := conn.Channel()
		if err != nil {
			p.close()
			return nil, err
		}

		// ExchangeDeclare is idempotent
		if err := ch.ExchangeDeclare(
			exCfg.Name,
			exCfg.Type,
			true,  // durable
			false, // auto-delete
			false, // internal
			false, // no-wait
			nil,
		); err != nil {
			ch.Close()
			p.close()
			return nil, fmt.Errorf("declare exchange failed: %w", err)
		}

		// Enable publisher confirm
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			p.close()
			return nil, fmt.Errorf("enable confirm mode failed: %w", err)
		}

		pc := &pubChannel{
			ch:       ch,
			confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 8)),
		}

		p.chPool <- pc
	}

	return p, nil
}

// Publish publishes a message and waits for broker confirm
func (p *Publisher) Publish( ctx context.Context, routingKey string, payload any, ) error {

	pc, err := p.acquire(ctx)
	if err != nil {
		return err
	}
	defer p.release(pc)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := pc.ch.PublishWithContext(
		ctx,
		p.cfg.Name,
		routingKey,
		false, // mandatory
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	); err != nil {
		return err
	}

	// Wait for broker confirm
	select {
	case confirm := <-pc.confirms:
		if !confirm.Ack {
			return errors.New("rabbitmq: publish not acknowledged by broker")
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (p *Publisher) acquire(ctx context.Context) (*pubChannel, error) {
	select {
	case pc, ok := <-p.chPool:
		if !ok || pc == nil {
			return nil, errors.New("rabbitmq: publisher closed")
		}
		return pc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *Publisher) release(pc *pubChannel) {
	if pc == nil {
		return
	}

	select {
	case p.chPool <- pc:
	default:
		pc.ch.Close()
	}
}

func (p *Publisher) Close() {
	p.once.Do(func() {
		p.close()
	})
}

func (p *Publisher) close() {
	close(p.chPool)
	for pc := range p.chPool {
		pc.ch.Close()
	}
}
