package cache

import (
	"context"
	"time"

	"air-social/pkg"
)

// TieredCache composes L1 (MemCache) and L2 (RedisStore) into a single lookup chain.
// Lookup order: L1 → L2 → loader (DB query).
// Backfill: a DB hit populates L2, then L1.
// Invalidation: Delete removes the key from both L1 and L2.
//
// l1 and l2 may each be nil to skip that layer (e.g. L2-only for auth).
type TieredCache[V any] struct {
	l1    Cache[V]
	l2    Cache[V]
	l1TTL time.Duration
	l2TTL time.Duration
}

// NewTieredCache creates a TieredCache. Pass nil for l1 or l2 to skip that layer.
func NewTieredCache[V any](l1, l2 Cache[V], l1TTL, l2TTL time.Duration) *TieredCache[V] {
	return &TieredCache[V]{l1: l1, l2: l2, l1TTL: l1TTL, l2TTL: l2TTL}
}

// GetOrLoad looks up the key in L1 → L2 → loader (DB), backfilling on each miss.
// loader is called only when both L1 and L2 miss. If loader returns an error,
// the error is propagated and nothing is cached.
func (c *TieredCache[V]) GetOrLoad(ctx context.Context, key string, loader func(context.Context) (V, error)) (V, error) {
	// L1 hit
	if c.l1 != nil {
		if val, err := c.l1.Get(ctx, key); err == nil {
			pkg.Log().Errorf("[tiered] key=%s -> L1 hit", key)
			return val, nil
		}
	}

	// L2 hit -> backfill L1
	if c.l2 != nil {
		if val, err := c.l2.Get(ctx, key); err == nil {
			pkg.Log().Errorf("[tiered] key=%s -> L2 hit, backfill L1", key)
			if c.l1 != nil {
				_ = c.l1.Set(ctx, key, val, c.l1TTL)
			}
			return val, nil
		}
	}

	// Both missed -> call loader (DB)
	pkg.Log().Errorf("[tiered] key=%s -> cache miss, calling loader", key)

	val, err := loader(ctx)
	if err != nil {
		pkg.Log().Errorf("[tiered] key=%s -> loader error: %v", key, err)
		return val, err
	}

	// Backfill L2 then L1
	if c.l2 != nil {
		_ = c.l2.Set(ctx, key, val, c.l2TTL)
	}
	if c.l1 != nil {
		_ = c.l1.Set(ctx, key, val, c.l1TTL)
	}
	pkg.Log().Errorf("[tiered] key=%s -> loader success, backfilled L2+L1", key)

	return val, nil
}

// Set writes the value to L2 then L1 using their configured TTLs.
// Use this to eagerly populate cache after a DB fetch (e.g. from GetByID/GetByEmail).
func (c *TieredCache[V]) Set(ctx context.Context, key string, val V) error {
	if c.l2 != nil {
		_ = c.l2.Set(ctx, key, val, c.l2TTL)
	}
	if c.l1 != nil {
		_ = c.l1.Set(ctx, key, val, c.l1TTL)
	}
	return nil
}

// Invalidate removes the key from both L1 and L2.
// Call this after a write (update/delete) to keep cache consistent.
func (c *TieredCache[V]) Invalidate(ctx context.Context, key string) error {
	if c.l1 != nil {
		_ = c.l1.Delete(ctx, key)
	}
	if c.l2 != nil {
		return c.l2.Delete(ctx, key)
	}
	return nil
}
