package email

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq"
)

type Worker struct {
	conn  *amqp.Connection
	cache domain.CacheStorage
	eCfg  rabbitmq.ExchangeConfig
	qCfg  rabbitmq.QueueConfig
	disp  domain.EventHandler

	ch   *amqp.Channel
	done chan struct{}
	once sync.Once
}

func NewEmailWorker(
	conn *amqp.Connection,
	cache domain.CacheStorage,
	disp domain.EventHandler,
	eCfg rabbitmq.ExchangeConfig,
	qCfg rabbitmq.QueueConfig,
) *Worker {
	return &Worker{
		conn:  conn,
		cache: cache,
		eCfg:  eCfg,
		qCfg:  qCfg,
		disp:  disp,
		done:  make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}

	if err := setupExchange(ch, w.eCfg); err != nil {
		ch.Close()
		return err
	}

	queueName, err := setupQueue(ch, w.qCfg)
	if err != nil {
		ch.Close()
		return err
	}

	if err := bindQueue(ch, queueName, w.eCfg, w.qCfg); err != nil {
		ch.Close()
		return err
	}

	if err := setupQos(ch); err != nil {
		ch.Close()
		return err
	}

	msgs, err := startConsume(ch, queueName)
	if err != nil {
		ch.Close()
		return err
	}

	w.ch = ch
	wg.Add(1)
	go consumeLoop(ctx, w.cache, msgs, w.disp, w.done, wg)

	return nil
}

func (w *Worker) Stop() error {
	w.once.Do(func() {
		close(w.done)
	})
	if w.ch != nil {
		return w.ch.Close()
	}
	return nil
}
