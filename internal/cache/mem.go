package cache

import (
	"context"
	"sync"
	"time"
)

type entry[V any] struct {
	val       V
	expiresAt time.Time // zero value = no expiration
}

func (e entry[V]) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

type MemCache[V any] struct {
	mu      sync.RWMutex
	items   map[string]entry[V]
	maxSize int // 0 = unlimited
}

// NewMemCache creates a new MemCache and starts the background janitor.
// maxSize: maximum number of entries (0 = unlimited).
// janitorInterval: how often expired entries are swept (e.g. 5*time.Minute).
func NewMemCache[V any](maxSize int, janitorInterval time.Duration) *MemCache[V] {
	c := &MemCache[V]{
		items:   make(map[string]entry[V]),
		maxSize: maxSize,
	}
	go c.janitor(janitorInterval)
	return c
}

// Get returns the cached value for the given key.
// Returns ErrCacheMiss if the key does not exist or has expired.
func (c *MemCache[V]) Get(_ context.Context, key string) (V, error) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()

	var zero V
	if !ok || e.expired() {
		return zero, ErrCacheMiss
	}
	return e.val, nil
}

// Set stores the value with the given TTL. TTL=0 means no expiration.
// If maxSize is set and the cache is full, the entry is silently dropped.
func (c *MemCache[V]) Set(_ context.Context, key string, val V, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.items) >= c.maxSize {
		// Cache is full: drop to avoid unbounded memory growth.
		// Upgrade to LRU eviction if needed in the future.
		return nil
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = entry[V]{val: val, expiresAt: expiresAt}
	return nil
}

// Delete removes the key from the cache. No error if the key does not exist.
func (c *MemCache[V]) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

// janitor runs in background and sweeps expired entries at the given interval.
func (c *MemCache[V]) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		for k, e := range c.items {
			if e.expired() {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
