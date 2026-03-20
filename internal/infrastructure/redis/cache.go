package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

// Redis data structures:
// - String: string / byte array
// - Hash: map (dictionary of field-value pairs)
// - List: linked list (ordered, push/pop)
// - Set: hash set (unique elements, no order)
// - ZSet: sorted set (skiplist for order + hash map for lookup)
type cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) (*cache, error) {
	if client == nil {
		return nil, errors.New("redis client cannot nil")
	}
	return &cache{client: client}, nil
}

// common.BasicCache

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

// common.HashCache

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

// common.SortedSetCache

func (r *cache) ZAdd(ctx context.Context, key string, score float64, member any) error {
	return r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (r *cache) ZRem(ctx context.Context, key string, members ...any) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

func (r *cache) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRevRange(ctx, key, start, stop).Result()
}

func (r *cache) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) error {
	return r.client.ZRemRangeByRank(ctx, key, start, stop).Err()
}

func (r *cache) ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	opt := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}
	return r.client.ZRevRangeByScore(ctx, key, opt).Result()
}

func (r *cache) Pipeline() common.CacheBatch {
	return &redisBatch{
		pipe: r.client.Pipeline(),
	}
}

// common.CacheBatch

type redisBatch struct {
	pipe redis.Pipeliner
}

func (b *redisBatch) ZAdd(key string, score float64, member any) {
	b.pipe.ZAdd(context.Background(), key, redis.Z{
		Score:  score,
		Member: member,
	})
}

func (b *redisBatch) ZRem(key string, members ...any) {
	b.pipe.ZRem(context.Background(), key, members...)
}

func (b *redisBatch) ZRemRangeByRank(key string, start, stop int64) {
	b.pipe.ZRemRangeByRank(context.Background(), key, start, stop)
}

func (b *redisBatch) Exec(ctx context.Context) error {
	_, err := b.pipe.Exec(ctx)
	return err
}
