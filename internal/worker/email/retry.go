package email

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/cache"
	"air-social/pkg"
)

const defaultMaxRetry = 3

func handleRetry(ctx context.Context, c cache.CacheStorage, msg amqp.Delivery, err error) {
	key := cache.GetEmailRetryKey(msg.MessageId)
	retry := getRetryCount(ctx, key, c)

	if retry < defaultMaxRetry {
		updateRetryCount(ctx, key, c, retry+1)
		time.Sleep(1 * time.Second)
		msg.Nack(false, true) // requeue
		return
	}

	pkg.Log().Errorw("processing failed, dropped message", "error", err, "retry", retry, "msg_id", msg.MessageId)
	msg.Nack(false, false)
	deleteRetryCount(ctx, key, c)
}
 
func getRetryCount(ctx context.Context, key string, c cache.CacheStorage) int {
	var retry int
	_ = c.Get(ctx, key, &retry)
	return retry
}

func updateRetryCount(ctx context.Context, key string, c cache.CacheStorage, retry int) error {
	return c.Set(ctx, key, retry, 24*time.Hour)
}

func deleteRetryCount(ctx context.Context, key string, c cache.CacheStorage) {
	_ = c.Delete(ctx, key)
}
