package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"air-social/pkg"
)

type redisCache struct {
	client *redis.Client
}

func newRedisCache(client *redis.Client) *redisCache {
	return &redisCache{client: client}
}

func (r *redisCache) Get(ctx context.Context, key string, dst any) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("cache: %w", pkg.ErrNotFound)
		}
		return err
	}
	return json.Unmarshal([]byte(data), dst)
}

func (r *redisCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, er := json.Marshal(val)
	if er != nil {
		return er
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCache) IsExist(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
