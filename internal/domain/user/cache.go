package user

import (
	"context"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
)

type Cache interface {
	Get(ctx context.Context, id int64, loader func(context.Context) (*UserSummary, error)) (*UserSummary, error)
	Invalidate(ctx context.Context, id int64) error
}

type userCache struct {
	store appcache.TieredStore[UserSummary]
}

func NewCache(store appcache.TieredStore[UserSummary]) *userCache {
	return &userCache{store: store}
}

func (c *userCache) Get(ctx context.Context, id int64, loader func(context.Context) (*UserSummary, error)) (*UserSummary, error) {
	val, err := c.store.GetOrLoad(ctx, userCacheKey(id), func(ctx context.Context) (UserSummary, error) {
		loaded, err := loader(ctx)
		if err != nil {
			return UserSummary{}, err
		}
		return *loaded, nil
	})
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (c *userCache) Invalidate(ctx context.Context, id int64) error {
	return c.store.Invalidate(ctx, userCacheKey(id))
}

func userCacheKey(id int64) string {
	return common.BuildCacheKey("user", "detail", id)
}
