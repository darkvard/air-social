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

// eventRoutingMap is the single source of truth that maps
// domain-level event types to infrastructure routing keys.
var eventRoutingMap = map[common.EventType]string{
	common.EventVerify:        config.EmailVerifyQueueConfig.RoutingKey,
	common.EventResetPassword: config.EmailResetPasswordQueueConfig.RoutingKey,
}

// pubChannel wraps an AMQP channel together with its confirm
// and return notification channels.
//
// Each channel operates in confirm mode so we can guarantee
// broker-level persistence acknowledgement.
type pubChannel struct {
	ch       *amqp.Channel
	confirms chan amqp.Confirmation
	returns  chan amqp.Return
}

// Publisher is a synchronous confirm-based RabbitMQ publisher.
//
// Design philosophy:
//
//	1 HTTP request = 1 publish = wait confirm = return channel.
//
// We intentionally keep the implementation SIMPLE.
// No self-healing, no auto-recreate channel, no force-close on timeout.
// Simplicity here increases reliability for request-response flows.
type Publisher struct {
	conn   *amqp.Connection
	cfg    config.ExchangeConfig
	chPool chan *pubChannel
	once   sync.Once
}

// NewPublisher initializes a pool of confirm-enabled channels.
//
// Exchange declaration is idempotent, so it is safe to call during startup.
// Channels are created once and reused to avoid channel churn.
func NewPublisher(conn *amqp.Connection, eCfg config.ExchangeConfig, poolSize int) (*Publisher, error) {
	if poolSize <= 0 {
		poolSize = 1
	}

	p := &Publisher{
		conn:   conn,
		cfg:    eCfg,
		chPool: make(chan *pubChannel, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		pc, err := p.createChannel()
		if err != nil {
			p.Close()
			return nil, err
		}
		p.chPool <- pc
	}

	return p, nil
}

// createChannel creates one AMQP channel configured in confirm mode.
func (p *Publisher) createChannel() (*pubChannel, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel failed: %w", err)
	}

	// Ensure exchange exists (idempotent operation)
	if err := topology.SetupExchange(ch, p.cfg); err != nil {
		ch.Close()
		return nil, err
	}

	// Enable publisher confirms.
	// Broker will ACK once message is durably persisted.
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("enable confirm failed: %w", err)
	}

	return &pubChannel{
		ch: ch,
		// NotifyPublish: receives ACKs/NACKs from the broker
		confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 10)),
		// NotifyReturn: receives messages that couldn't be routed (if mandatory=true)
		returns: ch.NotifyReturn(make(chan amqp.Return, 1)),
	}, nil
}

// Publish sends a domain event to RabbitMQ and waits synchronously
// for broker confirmation (ACK/NACK or Return).
//
// Flow:
//	acquire channel
//	resolve routing key
//	publish
//	wait confirm
//	release channel
func (p *Publisher) Publish(ctx context.Context, event common.Event) error {
	pc, err := p.acquire(ctx)
	if err != nil {
		return err
	}
	defer p.release(pc)

	routingKey, ok := eventRoutingMap[event.Typ]
	if !ok {
		return fmt.Errorf("unsupported event type: %s", event.Typ)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Capture the next publish sequence number.
	// We wait specifically for the confirm matching this sequence.
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

	// Wait for broker confirm or context cancellation.
	for {
		select {
		case ret := <-pc.returns:
			// If mandatory=true and routingKey doesn't match any queue, broker returns the msg
			return fmt.Errorf(
				"message returned: exchange=%s routingKey=%s reason=%s",
				ret.Exchange,
				ret.RoutingKey,
				ret.ReplyText,
			)

		case confirm, ok := <-pc.confirms:
			// Broker confirm (ACK/NACK)
			if !ok {
				return errors.New("confirm channel closed")
			}
			if confirm.DeliveryTag < seqNo {
				// DeliveryTag is a monotonically increasing counter per channel
				// Wait for the tag matching our current publish
				continue
			}
			if !confirm.Ack {
				return errors.New("message nacked by broker")
			}
			return nil

		case <-ctx.Done():
			// Context timeout or cancellation.
			//
			// IMPORTANT:
			// We DO NOT close the channel here.
			//
			// Reason:
			// - Confirm is asynchronous.
			// - Broker may still send ACK after our HTTP timeout.
			// - Closing the channel would:
			//     * discard pending confirms
			//     * break in-flight confirm stream
			//     * potentially destabilize connection
			//
			// Since channels are long-lived and reused,
			// we simply return the error and let the channel
			// continue operating normally.
			return ctx.Err()
		}
	}
}

// acquire gets a channel from the pool.
func (p *Publisher) acquire(ctx context.Context) (*pubChannel, error) {
	select {
	case pc, ok := <-p.chPool:
		if !ok || pc == nil {
			return nil, errors.New("publisher closed")
		}
		return pc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// release returns the channel back to the pool.
// If the pool is already full (rare case), the channel is closed.
func (p *Publisher) release(pc *pubChannel) {
	if pc == nil {
		return
	}

	select {
	case p.chPool <- pc:
	default:
		if !pc.ch.IsClosed() {
			pc.ch.Close()
		}
	}
}

// Close gracefully shuts down the publisher and all channels.
func (p *Publisher) Close() {
	p.once.Do(func() {
		close(p.chPool)
		for pc := range p.chPool {
			if pc != nil && !pc.ch.IsClosed() {
				pc.ch.Close()
			}
		}
	})
}
