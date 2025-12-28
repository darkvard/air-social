package email

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	cache cache.CacheStorage,
	msg amqp.Delivery,
	disp domain.EventHandler,
) {
	var evt domain.EventPayload
	if err := json.Unmarshal(msg.Body, &evt); err != nil {
		pkg.Log().Errorw("failed to unmarshal event", "error", err, "msg_id", msg.MessageId)
		msg.Nack(false, false)
		return
	}

	key := getCacheKey(msg.MessageId)
	if msg.MessageId != "" {
		exists, err := cache.IsExist(ctx, key)
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
		pkg.Log().Errorw("failed to handle event", "error", err, "type", evt.EventType)
		handleRetry(msg)
		return
	}

	if msg.MessageId != "" {
		_ = cache.Set(ctx, key, "1", 24*time.Hour)
	}

	msg.Ack(false)
}

func getCacheKey(id string) string {
	return fmt.Sprintf(cache.WorkerEmailProcessed+"%s", id)
}
