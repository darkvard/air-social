package cache

import (
	"context"
	"time"
)

// TieredStore is the interface for a multi-layer cache with loader fallback.
// Callers depend on this interface; TieredCache is the concrete implementation.
type TieredStore[V any] interface {
	// GetOrLoad looks up the key across cache layers, calling loader on a full miss.
	GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error)
	// Set writes the value into all cache layers using their configured TTLs.
	// Use this to eagerly populate the cache after a DB fetch.
	Set(ctx context.Context, key string, val V) error
	// Invalidate removes the key from all cache layers.
	Invalidate(ctx context.Context, key string) error
}

// Cache is a generic key-value cache interface.
// Implementations: MemCache (L1), RedisStore (L2).
// Use TieredCache to compose L1 + L2 + DB loader.
type Cache[V any] interface {
	// Get returns the cached value. Returns ErrCacheMiss if the key does not exist or has expired.
	Get(ctx context.Context, key string) (V, error)
	// Set stores the value with the given TTL. TTL=0 means no expiration.
	Set(ctx context.Context, key string, val V, ttl time.Duration) error
	// Delete removes the key. No error if the key does not exist.
	Delete(ctx context.Context, key string) error
}

// AtomicCache extends Cache with atomic operations required by auth flows.
// Use only for cases that need cross-instance consistency (no L1).
type AtomicCache[V any] interface {
	Cache[V]
	// SetNX stores the value only if the key does not already exist (atomic).
	// Returns true if the key was set, false if it already existed.
	SetNX(ctx context.Context, key string, val V, ttl time.Duration) (bool, error)
	// Exists reports whether the key is present in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}

// HashStore provides Redis Hash operations.
// Used by the stats write-behind cache (delta counters).
// Not generic: Redis Hash fields are always string, values are int64/string by protocol.
type HashStore interface {
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HMGet(ctx context.Context, key string, fields ...string) ([]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	// Eval executes a Lua script atomically on the Redis server.
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
}

// SortedSetStore provides Redis Sorted Set operations.
// Used by the feed newsfeed cache (cursor-paginated post IDs by timestamp score).
// Not generic: Redis ZSet members are always string, scores are always float64 by protocol.
type SortedSetStore interface {
	ZAdd(ctx context.Context, key string, score float64, member any) error
	ZRem(ctx context.Context, key string, members ...any) error
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) error
	// ZRevRangeByScore returns members with scores between min and max in descending order.
	// Used for cursor-based pagination: pass "(cursor" as max for exclusive lower bound.
	ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error)
	// Pipeline groups multiple write operations into a single network round-trip.
	Pipeline() Batch
}

// Batch is a grouped set of SortedSet write operations executed in a single round-trip.
type Batch interface {
	ZAdd(key string, score float64, member any)
	ZRem(key string, members ...any)
	ZRemRangeByRank(key string, start, stop int64)
	Exec(ctx context.Context) error
}
