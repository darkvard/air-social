package email

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
)

type Deps struct {
	Conn       *amqp.Connection
	Cache      common.Cache
	Dispatcher common.EventDispatcher
}

type Consumer struct {
	conn *amqp.Connection
	eCfg config.ExchangeConfig
	qCfg config.QueueConfig

	cache common.Cache
	disp  common.EventDispatcher

	ch   *amqp.Channel
	done chan struct{}
	once sync.Once
}

func NewConsumer(deps Deps, event common.EventType) *Consumer {
	var queueCfg config.QueueConfig

	switch event {
	case common.EventVerify:
		queueCfg = config.EmailVerifyQueueConfig
	case common.EventResetPassword:
		queueCfg = config.EmailResetPasswordQueueConfig
	default:
		pkg.Log().Error("consumer: invalid event type")
		return nil
	}

	return &Consumer{
		conn:  deps.Conn,
		eCfg:  config.TopicEventsExchange,
		qCfg:  queueCfg,
		cache: deps.Cache,
		disp:  deps.Dispatcher,
		done:  make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	pkg.Log().Infof("Consumer started for queue:", c.qCfg.Queue)

	ch, err := topology.PrepareConsumerChannel(c.conn, c.eCfg, c.qCfg)
	if err != nil {
		return err
	}

	msgs, err := topology.StartConsume(ch, c.qCfg.Queue)
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
	pkg.Log().Infof("Received message:", msg.RoutingKey, "message id", msg.MessageId)
	var event common.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		msg.Nack(false, false) // drop malformed messages
		return
	}

	processedKey := getProcessedKey(msg.MessageId)

	// LOCK: Use short lockExpiry (10m) to allow recovery if worker crashes
	ok, err := c.cache.SetNX(ctx, processedKey, processing, lockExpiry)
	if err != nil {
		msg.Nack(false, true) // requeue on cache error
		return
	}
	if !ok {
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
	msg.Ack(false)
}

func (c *Consumer) handleFailure(ctx context.Context, msg amqp.Delivery, err error) {
	// if the error won't be fixed by retrying, drop it immediately
	if pkg.IsPermanentError(err) {
		msg.Nack(false, false)
		return
	}

	retryKey := getRetryKey(msg.MessageId)
	var retryCount int
	_ = c.cache.Get(ctx, retryKey, &retryCount)

	if retryCount < defaultMaxRetry {
		// track retry attempts in cache
		_ = c.cache.Set(ctx, retryKey, retryCount+1, 1*time.Hour)
		// unlock: delete processedKey so the next retry can pass the SetNX check
		_ = c.cache.Delete(ctx, getProcessedKey(msg.MessageId))
		msg.Nack(false, true)
		return
	}

	// max retries reached: drop the message to avoid infinite loop
	msg.Nack(false, false)
}

func getProcessedKey(token string) string {
	return common.BuildCacheKey("worker", "email", "processed", token)
}

func getRetryKey(token string) string {
	return common.BuildCacheKey("worker", "email", "retry", token)
}
