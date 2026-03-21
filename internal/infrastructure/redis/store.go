package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	appcache "air-social/internal/cache"
)

// RedisStore is a generic L2 cache backed by Redis String keys with JSON encoding.
// It implements cache.AtomicCache[V] (superset of cache.Cache[V]).
// Keys are expected to be fully qualified (built via common.BuildCacheKey).
type RedisStore[V any] struct {
	client *redis.Client
}

func NewRedisStore[V any](client *redis.Client) *RedisStore[V] {
	return &RedisStore[V]{client: client}
}

// Get decodes the JSON value stored at the key into V.
// Returns appcache.ErrCacheMiss if the key does not exist or has expired.
func (s *RedisStore[V]) Get(ctx context.Context, key string) (V, error) {
	var zero V
	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return zero, appcache.ErrCacheMiss
		}
		return zero, err
	}
	if err := json.Unmarshal([]byte(data), &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

// Set encodes val as JSON and stores it with the given TTL. TTL=0 means no expiration.
func (s *RedisStore[V]) Set(ctx context.Context, key string, val V, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, b, ttl).Err()
}

// Delete removes the key. No error if the key does not exist.
func (s *RedisStore[V]) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// SetNX stores the value only if the key does not already exist (atomic).
// Returns true if the key was set, false if it already existed.
func (s *RedisStore[V]) SetNX(ctx context.Context, key string, val V, ttl time.Duration) (bool, error) {
	b, err := json.Marshal(val)
	if err != nil {
		return false, err
	}
	return s.client.SetNX(ctx, key, b, ttl).Result()
}

// Exists reports whether the key is present in the cache.
func (s *RedisStore[V]) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
