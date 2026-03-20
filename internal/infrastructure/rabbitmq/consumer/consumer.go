package consumer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain/common"
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/topology"
	"air-social/pkg"
)

const (
	defaultMaxRetry = 3
	lockExpiry      = 10 * time.Minute
	processedExpiry = 24 * time.Hour
)

const (
	processing = "processing"
	done       = "done"
	worker     = "worker"
	retry      = "retry"
	processed  = "processed"
)

type Domain string

type Deps struct {
	Conn       *amqp.Connection
	Cache      common.BasicCache
	Dispatcher common.EventDispatcher
	QueueCfg   config.QueueConfig
	Domain     Domain
}

type Consumer struct {
	conn        *amqp.Connection
	ExchangeCfg config.ExchangeConfig
	QueueCfg    config.QueueConfig
	cache       common.BasicCache
	disp        common.EventDispatcher
	domain      string

	ch   *amqp.Channel
	done chan struct{}
	once sync.Once
}

func NewConsumer(deps Deps) *Consumer {
	return &Consumer{
		conn:        deps.Conn,
		ExchangeCfg: config.TopicEventsExchange,
		QueueCfg:    deps.QueueCfg,
		cache:       deps.Cache,
		disp:        deps.Dispatcher,
		domain:      string(deps.Domain),
		done:        make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	ch, err := topology.PrepareConsumerChannel(c.conn, c.ExchangeCfg, c.QueueCfg)
	if err != nil {
		return err
	}

	msgs, err := topology.StartConsume(ch, c.QueueCfg.Queue)
	if err != nil {
		ch.Close()
		return err
	}

	c.ch = ch

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.done:
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				taskCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				c.handleMessage(taskCtx, msg)
				cancel()
			}
		}
	})

	return nil
}

func (c *Consumer) Stop() error {
	c.once.Do(func() { close(c.done) })
	if c.ch != nil {
		return c.ch.Close()
	}
	return nil
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	pkg.Log().Infof("Worker [%s]: Received message %s", c.domain, msg.MessageId)

	var event common.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		msg.Nack(false, false) // drop malformed messages
		return
	}

	processedKey := c.getProcessedKey(msg.MessageId)

	// LOCK: Use short lockExpiry (10m) to allow recovery if worker crashes
	ok, err := c.cache.SetNX(ctx, processedKey, processing, lockExpiry)
	if err != nil {
		msg.Nack(false, true) // requeue on cache error
		return
	}
	if !ok {
		pkg.Log().Infof("Worker [%s]: Message %s is already processing or done. Skipping.", c.domain, msg.MessageId)
		msg.Ack(false) //skip if already being handled or finished
		return
	}

	// business logic
	if err := c.disp.Dispatch(ctx, event); err != nil {
		c.handleFailure(ctx, msg, err)
		return
	}

	// DONE: set long processedExpiry (24h) to prevent duplicates permanently
	c.cache.Set(ctx, processedKey, done, processedExpiry)
	pkg.Log().Infof("Worker [%s]: Successfully processed message %s", c.domain, msg.MessageId)
	msg.Ack(false)
}

func (c *Consumer) handleFailure(ctx context.Context, msg amqp.Delivery, err error) {
	// if the error won't be fixed by retrying, drop it immediately
	if pkg.IsPermanentError(err) {
		msg.Nack(false, false)
		return
	}

	retryKey := c.getRetryKey(msg.MessageId)
	var retryCount int
	_ = c.cache.Get(ctx, retryKey, &retryCount)

	if retryCount < defaultMaxRetry {
		// track retry attempts in cache
		_ = c.cache.Set(ctx, retryKey, retryCount+1, 1*time.Hour)
		pkg.Log().Warnf("Worker [%s]: Message %s failed (attempt %d). Requeueing...", c.domain, msg.MessageId, retryCount+1)
		// unlock: delete processedKey so the next retry can pass the SetNX check
		_ = c.cache.Delete(ctx, c.getProcessedKey(msg.MessageId))
		msg.Nack(false, true)
		return
	}

	// max retries reached: drop the message to avoid infinite loop
	pkg.Log().Errorf("Worker [%s]: Message %s failed permanently after max retries.", c.domain, msg.MessageId)
	msg.Nack(false, false)
}

func (c *Consumer) getProcessedKey(token string) string {
	return common.BuildCacheKey(worker, c.domain, processed, token)
}

func (c *Consumer) getRetryKey(token string) string {
	return common.BuildCacheKey(worker, c.domain, retry, token)
}
