package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain/common"
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/topology"
)

// eventRoutingMap defines the source of truth for event-to-routing-key mapping
var eventRoutingMap = map[common.EventType]string{
	common.EventVerify:        "email.verify",
	common.EventResetPassword: "email.reset_password",
}

type pubChannel struct {
	ch       *amqp.Channel
	confirms chan amqp.Confirmation
	returns  chan amqp.Return
}

type Publisher struct {
	conn   *amqp.Connection
	cfg    config.ExchangeConfig
	chPool chan *pubChannel
	once   sync.Once
}

// NewPublisher initializes the publisher with a pool of pre-configured channels
func NewPublisher(conn *amqp.Connection, eCfg config.ExchangeConfig, poolSize int) (*Publisher, error) {
	if poolSize <= 0 {
		poolSize = 1
	}

	p := &Publisher{
		conn:   conn,
		cfg:    eCfg,
		chPool: make(chan *pubChannel, poolSize),
	}

	// fill the pool with ready-to-use channels
	for i := 0; i < poolSize; i++ {
		pc, err := p.createPoolChannel()
		if err != nil {
			p.Close()
			return nil, err
		}
		p.chPool <- pc
	}

	return p, nil
}

// createPoolChannel handles the internal technical setup for each channel in the pool
func (p *Publisher) createPoolChannel() (*pubChannel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 1. Ensure exchange exists on the broker (Idempotent)
	if err := topology.SetupExchange(ch, p.cfg); err != nil {
		ch.Close()
		return nil, err
	}

	// 2. Enable Publisher Confirms: Broker will send an ACK once message is safely stored
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to enable confirm mode: %w", err)
	}

	return &pubChannel{
		ch: ch,
		// NotifyPublish: receives ACKs/NACKs from the broker
		confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 10)),
		// NotifyReturn: receives messages that couldn't be routed (if mandatory=true)
		returns: ch.NotifyReturn(make(chan amqp.Return, 1)),
	}, nil
}

// Publish routes a domain event to RabbitMQ with guaranteed delivery checks
func (p *Publisher) Publish(ctx context.Context, event common.Event) error {
	pc, err := p.acquire(ctx)
	if err != nil {
		return err
	}
	defer p.release(pc)

	// Resolve the infrastructure routing key from domain event type
	routingKey, ok := eventRoutingMap[event.Typ]
	if !ok {
		return fmt.Errorf("unsupported event type: %s", event.Typ)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Track this specific publish sequence for the confirmation
	seqNo := pc.ch.GetNextPublishSeqNo()

	err = pc.ch.PublishWithContext(ctx, p.cfg.Name, routingKey,
		true,  // Mandatory: if true, broker returns message via pc.returns if no queue matches
		false, // Immediate: deprecated in modern RabbitMQ
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Persistent: write message to disk to survive crashes
			MessageId:    event.ID,        // Use domain event ID for end-to-end tracing
			Timestamp:    event.Timestamp,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	// Block until broker confirms receipt or context expires
	return p.waitConfirm(ctx, pc, seqNo)
}

// waitConfirm orchestrates the reliability checks (returns and confirms)
func (p *Publisher) waitConfirm(ctx context.Context, pc *pubChannel, seqNo uint64) error {
	for {
		select {
		case ret := <-pc.returns:
			// If mandatory=true and routingKey doesn't match any queue, broker returns the msg
			return fmt.Errorf("rabbitmq: message returned, no queue found for key [%s]", ret.RoutingKey)

		case confirm, ok := <-pc.confirms:
			if !ok {
				return errors.New("rabbitmq: confirmation channel closed")
			}
			// DeliveryTag is a monotonically increasing counter per channel
			if confirm.DeliveryTag < seqNo {
				continue // Wait for the tag matching our current publish
			}
			if !confirm.Ack {
				return errors.New("rabbitmq: message nacked by broker (storage failure)")
			}
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// acquire retrieves a channel from the pool
func (p *Publisher) acquire(ctx context.Context) (*pubChannel, error) {
	select {
	case pc, ok := <-p.chPool:
		if !ok || pc == nil {
			return nil, errors.New("publisher is closed")
		}
		return pc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// release returns a healthy channel back to the pool
func (p *Publisher) release(pc *pubChannel) {
	if pc != nil && !pc.ch.IsClosed() {
		p.chPool <- pc
	}
}

// Close gracefully shuts down all channels in the pool
func (p *Publisher) Close() {
	p.once.Do(func() {
		close(p.chPool)
		for pc := range p.chPool {
			if pc != nil && pc.ch != nil {
				pc.ch.Close()
			}
		}
	})
}
