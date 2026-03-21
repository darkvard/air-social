package usecase

import (
	"context"
	"time"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
)

const (
	summaryCacheL1TTL = 5 * time.Minute
	summaryCacheL2TTL = 12 * time.Hour
)

type Deps struct {
	Repo  user.Repository
	Cache appcache.TieredStore[*user.UserSummary]
	Link  common.LinkProvider
	Media MediaConfirmer
}

func GetKey(userID int64) string {
	return common.BuildCacheKey("user", "detail", userID)
}

func getUserCache(ctx context.Context, c appcache.TieredStore[*user.UserSummary], id int64, loader func(context.Context) (*user.UserSummary, error)) (*user.UserSummary, error) {
	return c.GetOrLoad(ctx, GetKey(id), loader)
}

func clearUserCache(ctx context.Context, c appcache.TieredStore[*user.UserSummary], id int64) error {
	return c.Invalidate(ctx, GetKey(id))
}
