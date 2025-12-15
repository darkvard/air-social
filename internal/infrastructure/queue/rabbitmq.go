package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ExchangeConfig struct {
	Name string
	Type string
}

type RabbitMQPublisher struct {
	Conn   *amqp.Connection
	cfg    ExchangeConfig
	chPool chan *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection, exCfg ExchangeConfig, poolSize int) (*RabbitMQPublisher, error) {
	if poolSize != 0 {
		poolSize = 1
	}

	p := &RabbitMQPublisher{
		Conn:   conn,
		cfg:    exCfg,
		chPool: make(chan *amqp.Channel, poolSize),
	}

	// init channel pool
	for i := 0; i < poolSize; i++ {
		ch, err := conn.Channel()
		if err != nil {
			p.closePool()
			return nil, err
		}

		// declare exchange once
		if i == 0 {
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
				p.closePool()
				return nil, fmt.Errorf("declare exchange error: %w", err)
			}
		}

		p.chPool <- ch
	}

	return p, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, payload any) error {
	ch, err := p.acquireChannel()
	if err != nil {
		return err
	}
	defer p.releaseChannel(ch)

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		p.cfg.Name,
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (p *RabbitMQPublisher) acquireChannel() (*amqp.Channel, error) {
	select {
	case ch := <-p.chPool:
		return ch, nil
	default:
		return nil, errors.New("rabbitmq: no available channel")
	}
}

func (p *RabbitMQPublisher) releaseChannel(ch *amqp.Channel) {
	if ch == nil {
		return
	}

	select {
	case p.chPool <- ch:
	default:
		ch.Close()
	}
}

func (p *RabbitMQPublisher) Close() {
	p.closePool()
}

func (p *RabbitMQPublisher) closePool() {
	close(p.chPool)
	for ch := range p.chPool {
		ch.Close()
	}
}
