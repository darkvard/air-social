package email

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/cache"
	"air-social/internal/domain"
)

const (
	dedupKeyPrefix = "email:processed:"
	dedupTTL       = 24 * time.Hour
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
		msg.Nack(false, false)
		return
	}

	key := getCacheKey(msg.MessageId)
	if msg.MessageId != "" {
		exists, _ := cache.IsExist(ctx, key)
		if exists {
			msg.Ack(false)
			return
		}
	}

	if err := disp.Handle(ctx, evt); err != nil {
		handleRetry(msg)
		return
	}

	if msg.MessageId != "" {
		_ = cache.Set(ctx, key, "1", dedupTTL)
	}

	msg.Ack(false)
}

func getCacheKey(id string) string {
	return dedupKeyPrefix + id
}
