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
	conn *amqp.Connection, eCfg mess.ExchangeConfig, qCfg mess.QueueConfig, disp domain.EventHandler,
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
	durable := true
	autoDelete, internal, noWait, exclusive := false, false, false, false

	if err := ch.ExchangeDeclare(w.eCfg.Name, w.eCfg.Type, durable, autoDelete, internal, noWait, nil); err != nil {
		return "", err
	}
	q, err := ch.QueueDeclare(w.qCfg.Queue, durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		return "", err
	}
	if err := ch.QueueBind(q.Name, w.qCfg.RoutingKey, w.eCfg.Name, noWait, nil); err != nil {
		return "", err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return "", err
	}

	return q.Name, nil
}

func (w *EmailWorker) startConsume(ch *amqp.Channel, name string) (<-chan amqp.Delivery, error) {
	autoAck, exclusive, noLocal, noWait := false, false, false, false
	return ch.Consume(name, "", autoAck, exclusive, noLocal, noWait, nil)
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
			w.handleMessage(ctx, msg)
		}
	}
}

func (w *EmailWorker) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var evt domain.EventPayload
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		msg.Nack(false, false)
		return
	}

	if err := w.disp.Handle(ctx, evt); err != nil {
		w.handleRetry(msg)
		return
	}

	msg.Ack(false)
}

func (w *EmailWorker) handleRetry(msg amqp.Delivery) {
	retryCount := GetRetryCount(msg)

	if retryCount < DefaultMaxRetry {
		IncrementRetry(&msg)
		msg.Nack(false, true) // requeue
		return
	}

	msg.Nack(false, false) // DLQ / drop
}
