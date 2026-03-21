package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisHashStore struct {
	client *redis.Client
}

func NewRedisHashStore(client *redis.Client) *RedisHashStore {
	return &RedisHashStore{client: client}
}

func (s *RedisHashStore) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return s.client.HIncrBy(ctx, key, field, incr).Result()
}

func (s *RedisHashStore) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.client.HGetAll(ctx, key).Result()
}

func (s *RedisHashStore) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	res, err := s.client.HMGet(ctx, key, fields...).Result()
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

func (s *RedisHashStore) HDel(ctx context.Context, key string, fields ...string) error {
	return s.client.HDel(ctx, key, fields...).Err()
}

func (s *RedisHashStore) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return s.client.Eval(ctx, script, keys, args).Result()
}
