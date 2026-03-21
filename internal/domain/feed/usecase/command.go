package usecase

import (
	"context"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
	"air-social/internal/domain/feed/cache"
)

type FollowFetcher interface {
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type CommandDeps struct {
	CacheProvider cache.Provider
	FollowFetcher FollowFetcher
	FollowerCache appcache.TieredStore[[]int64]
}

type commandUseCase struct {
	cacheProvider cache.Provider
	followFetcher FollowFetcher
	followerCache appcache.TieredStore[[]int64]
}

func NewCommandUseCase(deps CommandDeps) *commandUseCase {
	return &commandUseCase{
		cacheProvider: deps.CacheProvider,
		followFetcher: deps.FollowFetcher,
		followerCache: deps.FollowerCache,
	}
}

func (u *commandUseCase) getFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	key := common.BuildCacheKey("follow", "followers", userID)
	return u.followerCache.GetOrLoad(ctx, key, func(ctx context.Context) ([]int64, error) {
		return u.followFetcher.GetFollowerIDs(ctx, userID)
	})
}

func (u *commandUseCase) DistributePost(ctx context.Context, postID int64, authorID int64, timestamp int64) error {
	followerIDs, err := u.getFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}

	followerIDs = append(followerIDs, authorID)

	return u.cacheProvider.PushPostToFeeds(ctx, followerIDs, postID, float64(timestamp))
}

func (u *commandUseCase) RevokePost(ctx context.Context, postID int64, authorID int64) error {
	followerIDs, err := u.getFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}
	followerIDs = append(followerIDs, authorID)

	return u.cacheProvider.RemovePostFromFeeds(ctx, followerIDs, postID)
}
