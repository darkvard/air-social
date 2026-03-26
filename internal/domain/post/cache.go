package post

import (
	"context"
	"errors"
	"slices"
	"sync"

	"golang.org/x/sync/errgroup"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
)

type Cache interface {
	Get(ctx context.Context, id int64, loader func(context.Context) (*Post, error)) (*Post, error)

	GetBatch(ctx context.Context, ids []int64, batchLoader func(context.Context, []int64) ([]*Post, error)) ([]*Post, error)

	Invalidate(ctx context.Context, id int64) error
}

var errCacheMiss = errors.New("post: cache miss")

type postCache struct {
	store appcache.TieredStore[Post]
}

func NewCache(store appcache.TieredStore[Post]) Cache {
	return &postCache{store: store}
}

func (c *postCache) Get(ctx context.Context, id int64, loader func(context.Context) (*Post, error)) (*Post, error) {
	val, err := c.store.GetOrLoad(ctx, postCacheKey(id), func(ctx context.Context) (Post, error) {
		loaded, err := loader(ctx)
		if err != nil {
			return Post{}, err
		}
		return *loaded, nil
	})
	if err != nil {
		return nil, err
	}
	return &val, nil
}

// GetBatch returns multiple PostCores using a check-then-batch-load strategy:
//  1. Fan-out: check L1→L2 concurrently for each ID (sentinel loader, no DB call).
//  2. Collect hits and missIDs.
//  3. One batchLoader call for all misses (avoids N×1 queries common in feed pages).
//  4. Backfill cache for each DB result.
//  5. Return in original order, skipping nils (post deleted since feed was written).
func (c *postCache) GetBatch(
	ctx context.Context,
	ids []int64,
	batchLoader func(context.Context, []int64) ([]*Post, error),
) ([]*Post, error) {
	type slot struct {
		post *Post
		miss bool
	}
	slots := make([]slot, len(ids))

	var mu sync.Mutex
	var missIDs []int64

	g, gctx := errgroup.WithContext(ctx)
	for i, id := range ids {
		g.Go(func() error {
			val, err := c.store.GetOrLoad(gctx, postCacheKey(id), func(context.Context) (Post, error) {
				return Post{}, errCacheMiss
			})
			if errors.Is(err, errCacheMiss) {
				mu.Lock()
				slots[i].miss = true
				missIDs = append(missIDs, id)
				mu.Unlock()
				return nil
			}
			if err != nil {
				return err
			}
			slots[i].post = &val
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Sort for deterministic DB query — goroutine fan-out is non-deterministic.
	slices.Sort(missIDs)
	if len(missIDs) > 0 {
		loaded, err := batchLoader(ctx, missIDs)
		if err != nil {
			return nil, err
		}

		loadedMap := make(map[int64]*Post, len(loaded))
		for _, item := range loaded {
			_ = c.store.Set(ctx, postCacheKey(item.ID), *item)
			loadedMap[item.ID] = item
		}

		for i, id := range ids {
			if slots[i].miss {
				slots[i].post = loadedMap[id] // nil if post was deleted since feed was written
			}
		}
	}

	result := make([]*Post, 0, len(ids))
	for _, s := range slots {
		if s.post != nil {
			result = append(result, s.post)
		}
	}
	return result, nil
}

// Invalidate removes the PostCore from both L1 and L2.
func (c *postCache) Invalidate(ctx context.Context, id int64) error {
	return c.store.Invalidate(ctx, postCacheKey(id))
}

func postCacheKey(id int64) string {
	return common.BuildCacheKey("post", "core", id)
}
