package message

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	unreadTTL    = 30 * 24 * time.Hour
	unreadKeyFmt = "chat:unread:%d"
)

type UnreadStore struct {
	client *redis.Client
}

func NewUnreadStorage(client *redis.Client) *UnreadStore {
	return &UnreadStore{client: client}
}

func (r *UnreadStore) Increment(ctx context.Context, userID int64, convID string) error {
	key := r.getKey(userID)
	pipe := r.client.Pipeline()
	pipe.HIncrBy(ctx, key, convID, 1)
	pipe.Expire(ctx, key, unreadTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *UnreadStore) Reset(ctx context.Context, userID int64, convID string) error {
	key := r.getKey(userID)
	return r.client.HDel(ctx, key, convID).Err()
}

func (r *UnreadStore) Get(ctx context.Context, userID int64, convID string) (int64, error) {
	key := r.getKey(userID)
	val, err := r.client.HGet(ctx, key, convID).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return val, err
}

func (r *UnreadStore) GetAll(ctx context.Context, userID int64) (map[string]int64, error) {
	key := r.getKey(userID)
	raw, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]int64, len(raw))
	for convID, v := range raw {
		result[convID], _ = strconv.ParseInt(v, 10, 64)
	}
	return result, nil
}

func (r *UnreadStore) getKey(userID int64) string {
	return fmt.Sprintf(unreadKeyFmt, userID)
}
