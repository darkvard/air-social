package usecase

import (
	"context"
	"time"

	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
)

const summaryCacheTTL = 12 * time.Hour

type Deps struct {
	Repo  user.Repository
	Cache common.BasicCache
	Link  common.LinkProvider
	Media MediaConfirmer
}

func GetKey(userID int64) string {
	return common.BuildCacheKey("user", "info", "public", userID)
}

func setUserCache(ctx context.Context, cache common.BasicCache, user *user.UserSummary) error {
	key := GetKey(user.ID)
	return cache.Set(ctx, key, user, summaryCacheTTL)
}

func getUserCache(ctx context.Context, cache common.BasicCache, id int64) (*user.UserSummary, error) {
	key := GetKey(id)
	var cached user.UserSummary
	if err := cache.Get(ctx, key, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func clearUserCache(ctx context.Context, cache common.BasicCache, id int64) error {
	key := GetKey(id)
	return cache.Delete(ctx, key)
}
