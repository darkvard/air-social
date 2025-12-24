package worker

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	mess "air-social/internal/infrastructure/messaging"
)

type EmailWorker struct {
	conn *amqp.Connection
	eCfg mess.ExchangeConfig
	qCfg mess.QueueConfig
	disp domain.EventHandler
	ch   *amqp.Channel
	done chan struct{}
	once sync.Once
}

func NewEmailWorker(
	conn *amqp.Connection,
	eCfg mess.ExchangeConfig,
	qCfg mess.QueueConfig,
	disp domain.EventHandler,
) *EmailWorker {
	return &EmailWorker{
		conn: conn,
		eCfg: eCfg,
		qCfg: qCfg,
		disp: disp,
		done: make(chan struct{}),
	}
}

func (w *EmailWorker) Start(ctx context.Context, wg *sync.WaitGroup) (err error) {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			ch.Close()
		}
	}()

	queueName, err := w.ensureQueueTopology(ch)
	if err != nil {
		return err
	}

	msgs, err := w.startConsume(ch, queueName)
	if err != nil {
		return err
	}

	w.ch = ch
	wg.Add(1)
	go w.consumeLoop(ctx, msgs, wg)

	return nil
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

func (w *EmailWorker) openChannel() (*amqp.Channel, error) {
	ch, err := w.conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (w *EmailWorker) ensureQueueTopology(ch *amqp.Channel) (string, error) {
	if err := ch.ExchangeDeclare(
		w.eCfg.Name,
		w.eCfg.Type,
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return "", err
	}

	q, err := ch.QueueDeclare(
		w.qCfg.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return "", err
	}

	if err := ch.QueueBind(
		q.Name,
		w.qCfg.RoutingKey,
		w.eCfg.Name,
		false, // no-wait
		nil,
	); err != nil {
		return "", err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return "", err
	}

	return q.Name, nil
}

func (w *EmailWorker) startConsume(ch *amqp.Channel, name string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		name,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
}

func (w *EmailWorker) consumeLoop(ctx context.Context, msgs <-chan amqp.Delivery, wg *sync.WaitGroup) {
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
			if err := w.disp.Handle(ctx, evt); err != nil {
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}
}
