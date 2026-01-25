package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type pubChannel struct {
	ch       *amqp.Channel
	confirms chan amqp.Confirmation
	returns  chan amqp.Return
}

type Publisher struct {
	conn   *amqp.Connection
	cfg    ExchangeConfig
	chPool chan *pubChannel
	once   sync.Once
}

func NewPublisher(conn *amqp.Connection, eCfg ExchangeConfig, poolSize int) (*Publisher, error) {
	if poolSize <= 0 {
		poolSize = 1
	}

	p := &Publisher{
		conn:   conn,
		cfg:    eCfg,
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
			eCfg.Name,
			eCfg.Type,
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
			returns:  ch.NotifyReturn(make(chan amqp.Return, 1)),
		}

		p.chPool <- pc
	}

	return p, nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, payload any) error {
	pc, err := p.acquire(ctx)
	if err != nil {
		return err
	}
	defer p.release(pc)

	seqNo := pc.ch.GetNextPublishSeqNo()

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := pc.ch.PublishWithContext(
		ctx,
		p.cfg.Name,
		routingKey,
		true, // mandatory
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.NewString(),
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	); err != nil {
		return err
	}

	// Wait for broker confirm matching our sequence number
	for {
		select {
		case ret := <-pc.returns: // Routing fail
			return fmt.Errorf(
				"publish return: exchange = %s, routingKey = %s, reason = %s",
				ret.Exchange,
				ret.RoutingKey,
				ret.ReplyText,
			)

		case confirm, ok := <-pc.confirms: // Broker confirm
			if !ok {
				return errors.New("rabbitmq: confirm channel closed")
			}
			if confirm.DeliveryTag < seqNo {
				continue
			}
			if !confirm.Ack {
				return errors.New("rabbitmq: publish not acknowledged by broker")
			}
			return nil
		case <-ctx.Done(): // Timeout, cancel
			return ctx.Err()
		}
	}
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
