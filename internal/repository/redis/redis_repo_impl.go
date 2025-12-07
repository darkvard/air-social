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

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string, dst any) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("cache: %w", pkg.ErrKeyNotFound)
		}
		return err
	}
	return json.Unmarshal([]byte(data), dst)
}

func (r *RedisCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, er := json.Marshal(val)
	if er != nil {
		return er
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) IsExist(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
