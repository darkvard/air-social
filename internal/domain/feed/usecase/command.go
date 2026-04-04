package usecase

import (
	"context"

	"air-social/internal/cache"
	"air-social/internal/domain/feed"
	"air-social/internal/domain/follow"
)

type FollowFetcher interface {
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type CommandDeps struct {
	CacheProvider feed.Cache
	FollowFetcher FollowFetcher
	FollowerCache cache.TieredStore[[]int64]
}

type commandUseCase struct {
	deps CommandDeps
}

func NewCommandUseCase(deps CommandDeps) *commandUseCase {
	return &commandUseCase{deps: deps}
}

func (u *commandUseCase) getFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := follow.GetFollowCacheKey(userID)
	return u.deps.FollowerCache.GetOrLoad(ctx, key, func(ctx context.Context) ([]int64, error) {
		return u.deps.FollowFetcher.GetFollowerIDs(ctx, userID)
	})
}

func (u *commandUseCase) DistributePost(ctx context.Context, postID int64, authorID int64, timestamp int64) error {
	followerIDs, err := u.getFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}

	followerIDs = append(followerIDs, authorID)

	return u.deps.CacheProvider.PushPostToFeeds(ctx, followerIDs, postID, float64(timestamp))
}

func (u *commandUseCase) RevokePost(ctx context.Context, postID int64, authorID int64) error {
	followerIDs, err := u.getFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}
	followerIDs = append(followerIDs, authorID)

	return u.deps.CacheProvider.RemovePostFromFeeds(ctx, followerIDs, postID)
}
