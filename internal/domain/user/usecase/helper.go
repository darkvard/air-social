package usecase

import (
	"context"
	"time"

	"air-social/internal/domain"
	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
)

const summaryCacheTTL = 12 * time.Hour

type Deps struct {
	Repo  user.Repository
	Cache shared.Cache
	Link  shared.LinkProvider
	Media MediaConfirmer
}


func getKey(userID int64) string {
	return shared.BuildCacheKey("user", "info", "public", userID)
}

func setUserCache(ctx context.Context, cache domain.CacheStorage, user *user.UserSummary) error {
	key := getKey(user.ID)
	return cache.Set(ctx, key, user, summaryCacheTTL)
}

func getUserCache(ctx context.Context, cache domain.CacheStorage, id int64) (*user.UserSummary, error) {
	key := getKey(id)
	var cached user.UserSummary
	if err := cache.Get(ctx, key, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func clearUserCache(ctx context.Context, cache domain.CacheStorage, id int64) error {
	key := getKey(id)
	return cache.Delete(ctx, key)
}
