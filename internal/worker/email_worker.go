package worker

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	"air-social/internal/event"
	"air-social/internal/infrastructure/messaging"
)

type EmailWorker struct {
	conn  *amqp.Connection
	exCfg messaging.ExchangeConfig
	quCfg messaging.QueueConfig
	disp  event.EmailDispatcher
	ch    *amqp.Channel
	done  chan struct{}
	once sync.Once
}

func NewEmailWorker(
	conn *amqp.Connection,
	exCfg messaging.ExchangeConfig,
	quCfg messaging.QueueConfig,

	dispatcher event.EmailDispatcher,
) *EmailWorker {
	return &EmailWorker{
		conn:  conn,
		exCfg: exCfg,
		quCfg: quCfg,
		disp:  dispatcher,
		done:  make(chan struct{}),
	}
}

func (w *EmailWorker) Start(ctx context.Context, wg *sync.WaitGroup) error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}
	w.ch = ch

	if err := ch.ExchangeDeclare(
		w.exCfg.Name,
		w.exCfg.Type,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		w.quCfg.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	if err := ch.QueueBind(
		q.Name,
		w.quCfg.RoutingKey,
		w.exCfg.Name,
		false, // no-wait
		nil,
	); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	wg.Add(1)
	go w.loop(ctx, msgs, wg)

	return nil
}

func (w *EmailWorker) loop(
	ctx context.Context,
	msgs <-chan amqp.Delivery,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case <-w.done:
			return

		case msg, ok := <-msgs:
			if !ok {
				return
			}

			var evt domain.EventPayload
			if err := json.Unmarshal(msg.Body, &evt); err != nil {
				msg.Nack(false, false)
				continue
			}

			if err := w.disp.Dispatch(ctx, evt); err != nil {
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}
}

func (w *EmailWorker) Stop() error {
	w.once.Do(func() {
		close(w.done)
	})
	if w.ch != nil {
		return w.ch.Close()
	}
	return nil
}

