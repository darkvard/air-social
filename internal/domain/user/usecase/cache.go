package usecase

import (
	"context"
	"time"

	"air-social/internal/domain"
	"air-social/internal/domain/user"
)

const summaryCacheTTL = 12 * time.Hour

func setUserCache(ctx context.Context, cache domain.CacheStorage, user *user.UserSummaryResult) error {
	key := domain.GetUserSummaryKey(user.ID)
	return cache.Set(ctx, key, user, summaryCacheTTL)
}

func getUserCache(ctx context.Context, cache domain.CacheStorage, id int64) (*user.UserSummaryResult, error) {
	key := domain.GetUserSummaryKey(id)
	var cached user.UserSummaryResult
	if err := cache.Get(ctx, key, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func clearUserCache(ctx context.Context, cache domain.CacheStorage, id int64) error {
	key := domain.GetUserSummaryKey(id)
	return cache.Delete(ctx, key)
}
