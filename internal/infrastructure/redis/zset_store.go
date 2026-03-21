package redis

import (
	"context"

	"github.com/redis/go-redis/v9"

	appcache "air-social/internal/cache"
)

type redisBatch struct {
	pipe redis.Pipeliner
}

func (b *redisBatch) ZAdd(key string, score float64, member any) {
	b.pipe.ZAdd(context.Background(), key, redis.Z{Score: score, Member: member})
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

// RedisSortedSetStore implements cache.SortedSetStore using Redis Sorted Set data structure.
// Used by the feed newsfeed cache (cursor-paginated post IDs ordered by timestamp score).
type RedisSortedSetStore struct {
	client *redis.Client
}

// NewRedisSortedSetStore creates a new RedisSortedSetStore.
func NewRedisSortedSetStore(client *redis.Client) *RedisSortedSetStore {
	return &RedisSortedSetStore{client: client}
}

func (s *RedisSortedSetStore) ZAdd(ctx context.Context, key string, score float64, member any) error {
	return s.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

func (s *RedisSortedSetStore) ZRem(ctx context.Context, key string, members ...any) error {
	return s.client.ZRem(ctx, key, members...).Err()
}

func (s *RedisSortedSetStore) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return s.client.ZRevRange(ctx, key, start, stop).Result()
}

func (s *RedisSortedSetStore) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) error {
	return s.client.ZRemRangeByRank(ctx, key, start, stop).Err()
}

func (s *RedisSortedSetStore) ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	return s.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min, Max: max, Offset: offset, Count: count,
	}).Result()
}

func (s *RedisSortedSetStore) Pipeline() appcache.Batch {
	return &redisBatch{pipe: s.client.Pipeline()}
}
