package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/model"
	"air-social/internal/infrastructure/rabbitmq/topology"
	"air-social/pkg"
)

// todo: remove
type EmailWorker struct {
	conn  *amqp.Connection
	cache domain.CacheStorage
	eCfg  config.ExchangeConfig
	qCfg  config.QueueConfig
	disp  domain.EventDispatcher

	ch   *amqp.Channel
	done chan struct{}
	once sync.Once
}

func NewEmailWorker(
	conn *amqp.Connection,
	cache domain.CacheStorage,
	disp domain.EventDispatcher,
	eCfg config.ExchangeConfig,
	qCfg config.QueueConfig,
) *EmailWorker {
	return &EmailWorker{
		conn:  conn,
		cache: cache,
		eCfg:  eCfg,
		qCfg:  qCfg,
		disp:  disp,
		done:  make(chan struct{}),
	}
}

func (w *EmailWorker) Start(ctx context.Context, wg *sync.WaitGroup) error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}

	if err := topology.SetupExchange(ch, w.eCfg); err != nil {
		ch.Close()
		return err
	}

	queueName, err := topology.SetupQueue(ch, w.qCfg)
	if err != nil {
		ch.Close()
		return err
	}

	if err := topology.BindQueue(ch, queueName, w.eCfg, w.qCfg); err != nil {
		ch.Close()
		return err
	}

	if err := topology.SetupQos(ch); err != nil {
		ch.Close()
		return err
	}

	msgs, err := topology.StartConsume(ch, queueName)
	if err != nil {
		ch.Close()
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

func (w *EmailWorker) consumeLoop(
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
			w.handleMessage(ctx, msg)
		}
	}
}

func (w *EmailWorker) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var modelEvt model.Event
	if err := json.Unmarshal(msg.Body, &modelEvt); err != nil {
		pkg.Log().Errorw("failed to unmarshal event", "error", err, "msg_id", msg.MessageId)
		msg.Nack(false, false)
		return
	}

	evt := modelEvt.ToDomainEvent()

	key := domain.GetEmailProcessedKey(msg.MessageId)
	if msg.MessageId != "" {
		exists, err := w.cache.IsExist(ctx, key)
		if err != nil {
			pkg.Log().Warnw("failed to check idempotency key", "error", err, "msg_id", msg.MessageId)
		}
		if exists {
			pkg.Log().Infow("message already processed, skipping", "msg_id", msg.MessageId, "type", evt.EventType)
			msg.Ack(false)
			return
		}
	}

	if err := w.disp.Dispatch(ctx, evt); err != nil {
		if pkg.IsPermanentError(err) {
			pkg.Log().Errorw("permanent error detected, dropping", "error", err, "msg_id", msg.MessageId)
			msg.Nack(false, false)
			_ = w.cache.Delete(ctx, domain.GetEmailRetryKey(msg.MessageId))
		} else {
			w.handleRetry(ctx, msg, err)
		}
		return
	}
	if msg.MessageId != "" {
		_ = w.cache.Set(ctx, key, "1", 24*time.Hour)
	}

	msg.Ack(false)
}

const defaultMaxRetry = 3

func (w *EmailWorker) handleRetry(ctx context.Context, msg amqp.Delivery, err error) {
	key := domain.GetEmailRetryKey(msg.MessageId)
	var retry int
	_ = w.cache.Get(ctx, key, &retry)

	if retry < defaultMaxRetry {
		_ = w.cache.Set(ctx, key, retry+1, 24*time.Hour)
		time.Sleep(1 * time.Second)
		msg.Nack(false, true) // requeue
		return
	}

	pkg.Log().Errorw("processing failed, dropped message", "error", err, "retry", retry, "msg_id", msg.MessageId)
	msg.Nack(false, false)
	_ = w.cache.Delete(ctx, key)
}
