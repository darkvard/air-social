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

type cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) (*cache, error) {
	if client == nil {
		return nil, errors.New("redis client cannot nil")
	}
	return &cache{client: client}, nil
}

func (r *cache) Get(ctx context.Context, key string, dst any) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("cache: %w", pkg.ErrNotFound)
		}
		return err
	}
	return json.Unmarshal([]byte(data), dst)
}

func (r *cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *cache) SetNX(ctx context.Context, key string, val any, ttl time.Duration) (bool, error) {
	b, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	return r.client.SetNX(ctx, key, b, ttl).Result()
}

func (r *cache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *cache) IsExist(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (r *cache) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return r.client.HIncrBy(ctx, key, field, incr).Result()
}

func (r *cache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *cache) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	res, err := r.client.HMGet(ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}

	out := make([]string, len(res))
	for i, v := range res {
		if v == nil {
			continue
		}

		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("redis: unexpected type %T", v)
		}
		out[i] = s
	}
	return out, nil
}

func (r *cache) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *cache) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return r.client.Eval(ctx, script, keys, args).Result()
}
