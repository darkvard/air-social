package email

import (
	"context"
	"encoding/json"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/cache"
	"air-social/internal/domain"
	"air-social/pkg"
)

func consumeLoop(
	ctx context.Context,
	cache cache.CacheStorage,
	msgs <-chan amqp.Delivery,
	disp domain.EventHandler,
	done <-chan struct{},
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			handleMessage(ctx, cache, msg, disp)
		}
	}
}

func handleMessage(
	ctx context.Context,
	c cache.CacheStorage,
	msg amqp.Delivery,
	disp domain.EventHandler,
) {
	var evt domain.EventPayload
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		pkg.Log().Errorw("failed to unmarshal event", "error", err, "msg_id", msg.MessageId)
		msg.Nack(false, false)
		return
	}

	key := cache.GetEmailProcessedKey(msg.MessageId)
	if msg.MessageId != "" {
		exists, err := c.IsExist(ctx, key)
		if err != nil {
			pkg.Log().Warnw("failed to check idempotency key", "error", err, "msg_id", msg.MessageId)
		}
		if exists {
			pkg.Log().Infow("message already processed, skipping", "msg_id", msg.MessageId, "type", evt.EventType)
			msg.Ack(false)
			return
		}
	}

	if err := disp.Handle(ctx, evt); err != nil {
		if pkg.IsPermanentError(err) {
			pkg.Log().Errorw("permanent error detected, dropping", "error", err, "msg_id", msg.MessageId)
			msg.Nack(false, false)
			deleteRetryCount(ctx, cache.GetEmailRetryKey(msg.MessageId), c)
		} else {
			handleRetry(ctx, c, msg, err)
		}
		return
	}
	if msg.MessageId != "" {
		_ = c.Set(ctx, key, "1", cache.OneDayTime)
	}

	msg.Ack(false)
}
