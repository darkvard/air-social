package usecase

import (
	"context"

	"air-social/internal/domain/feed/cache"
)

type FollowFetcher interface {
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type CommandDeps struct {
	CacheProvider cache.Provider
	FollowFetcher FollowFetcher
}

type commandUseCase struct {
	cacheProvider cache.Provider
	followFetcher FollowFetcher
}

func NewCommandUseCase(deps CommandDeps) *commandUseCase {
	return &commandUseCase{
		cacheProvider: deps.CacheProvider,
		followFetcher: deps.FollowFetcher,
	}
}

func (u *commandUseCase) DistributePost(ctx context.Context, postID int64, authorID int64, timestamp int64) error {
	followerIDs, err := u.followFetcher.GetFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}

	followerIDs = append(followerIDs, authorID)

	err = u.cacheProvider.PushPostToFeeds(ctx, followerIDs, postID, float64(timestamp))
	if err != nil {
		return err
	}

	return nil
}

func (u *commandUseCase) RevokePost(ctx context.Context, postID int64, authorID int64) error {
	followerIDs, err := u.followFetcher.GetFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}
	followerIDs = append(followerIDs, authorID)

	return u.cacheProvider.RemovePostFromFeeds(ctx, followerIDs, postID)
}
