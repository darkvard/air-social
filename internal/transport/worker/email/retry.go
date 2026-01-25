package email

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/domain"
	"air-social/pkg"
)

const defaultMaxRetry = 3

func handleRetry(ctx context.Context, cache domain.CacheStorage, msg amqp.Delivery, err error) {
	key := domain.GetEmailRetryKey(msg.MessageId)
	retry := getRetryCount(ctx, key, cache)

	if retry < defaultMaxRetry {
		updateRetryCount(ctx, key, cache, retry+1)
		time.Sleep(1 * time.Second)
		msg.Nack(false, true) // requeue
		return
	}

	pkg.Log().Errorw("processing failed, dropped message", "error", err, "retry", retry, "msg_id", msg.MessageId)
	msg.Nack(false, false)
	deleteRetryCount(ctx, key, cache)
}

func getRetryCount(ctx context.Context, key string, cache domain.CacheStorage) int {
	var retry int
	_ = cache.Get(ctx, key, &retry)
	return retry
}

func updateRetryCount(ctx context.Context, key string, cache domain.CacheStorage, retry int) error {
	return cache.Set(ctx, key, retry, 24*time.Hour)
}

func deleteRetryCount(ctx context.Context, key string, cache domain.CacheStorage) {
	_ = cache.Delete(ctx, key)
}
