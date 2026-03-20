package common

import (
	"context"
	"fmt"
	"time"
)

// Cache patterns:
// - Set/Get: basic caching
// - SetNX: distributed lock / prevent duplicate
// - H*: object storage + counters
// - Z*: ranking / priority / scheduling
// - Eval: complex logic with atomic execution

type Cache interface {
	BasicCache
	HashCache
	SortedSetCache
}

type BasicCache interface {
	// Get value by key and decode into dst
	Get(ctx context.Context, key string, dst any) error
	// Set value with TTL (0 means no expiration)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	// Set value only if key does not exist (atomic)
	SetNX(ctx context.Context, key string, val any, ttl time.Duration) (bool, error)
	// Delete key from cache
	Delete(ctx context.Context, key string) error
	// Check if key exists
	IsExist(ctx context.Context, key string) (bool, error)
}

type HashCache interface {
	// Increment hash field by given value
	HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error)
	// Get all fields and values of a hash
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	// Get multiple fields from a hash
	HMGet(ctx context.Context, key string, fields ...string) ([]string, error)
	// Delete one or more fields from a hash
	HDel(ctx context.Context, key string, fields ...string) error
	// Execute Lua script atomically
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)
}

type SortedSetCache interface {
	// Add member to sorted set with score
	ZAdd(ctx context.Context, key string, score float64, member any) error
	// Remove member(s) from sorted set
	ZRem(ctx context.Context, key string, members ...any) error
	// Get members in descending order by score
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	// Remove members by rank range
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) error
	// Get members in descending order by score (using score values like timestamp). Useful for Keyset/Cursor Pagination.
	ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error)
	// Pipeline initializes a batch execution context to group multiple operations
	// into a single network request, significantly reducing latency.
	Pipeline() CacheBatch
}

// CacheBatch represents a collection of caching operations to be executed atomically or in bulk.
// It abstracts away the underlying infrastructure's pipeline mechanism.
type CacheBatch interface {
	// Queue a ZAdd operation in the batch
	ZAdd(key string, score float64, member any)
	// Queue a ZRem operation in the batch
	ZRem(key string, members ...any)
	// Queue a ZRemRangeByRank operation in the batch
	ZRemRangeByRank(key string, start, stop int64)
	// Execute all queued operations in a single network round-trip
	Exec(ctx context.Context) error
}

// <system>:<feature>:<state>:<id>
func BuildCacheKey(system, feature, state string, id any) string {
	return fmt.Sprintf("%s:%s:%s:%v", system, feature, state, id)
}
