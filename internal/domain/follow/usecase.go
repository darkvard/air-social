package follow

import (
	"context"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type UseCase interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowings(ctx context.Context, params GetFollowsParams) (common.OffsetPaginatedResult[FollowUser], error)
	GetFollowers(ctx context.Context, params GetFollowsParams) (common.OffsetPaginatedResult[FollowUser], error)
	GetRelationship(ctx context.Context, userID, targetID int64) (Relationship, error)
	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}

type UserFetcher interface {
	GetSummary(ctx context.Context, id int64) (*user.UserSummary, error)
}

type CacheInvalidator interface {
	Invalidate(ctx context.Context, key string) error
}

type Deps struct {
	FollowRepo       Repository
	UserFetcher      UserFetcher
	CacheInvalidator CacheInvalidator
}

type usecase struct {
	followRepo       Repository
	userFetcher      UserFetcher
	cacheInvalidator CacheInvalidator
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		followRepo:       deps.FollowRepo,
		userFetcher:      deps.UserFetcher,
		cacheInvalidator: deps.CacheInvalidator,
	}
}

func (u *usecase) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	if err := u.validateFollow(ctx, followerID, followeeID); err != nil {
		return err
	}
	if err := pkg.OrInternalError(u.followRepo.Create(ctx, followerID, followeeID)); err != nil {
		return err
	}
	if u.cacheInvalidator != nil {
		_ = u.cacheInvalidator.Invalidate(ctx, GetFollowCacheKey(followeeID))
	}
	return nil
}

func (u *usecase) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	if err := u.validateFollow(ctx, followerID, followeeID); err != nil {
		return err
	}
	if err := pkg.OrInternalError(u.followRepo.Delete(ctx, followerID, followeeID)); err != nil {
		return err
	}
	if u.cacheInvalidator != nil {
		_ = u.cacheInvalidator.Invalidate(ctx, GetFollowCacheKey(followeeID))
	}
	return nil
}

func (u *usecase) GetFollowings(ctx context.Context, params GetFollowsParams) (common.OffsetPaginatedResult[FollowUser], error) {
	params.Paging.NormalizePagination()

	data, total, err := u.followRepo.GetFollowings(ctx, params)
	if err != nil {
		return common.OffsetPaginatedResult[FollowUser]{}, err
	}

	return common.NewOffsetPaginatedResult(data, total, params.Paging.Page, params.Paging.Limit), nil
}

func (u *usecase) GetFollowers(ctx context.Context, params GetFollowsParams) (common.OffsetPaginatedResult[FollowUser], error) {
	params.Paging.NormalizePagination()

	data, total, err := u.followRepo.GetFollowers(ctx, params)
	if err != nil {
		return common.OffsetPaginatedResult[FollowUser]{}, err
	}

	return common.NewOffsetPaginatedResult(data, total, params.Paging.Page, params.Paging.Limit), nil
}

func (u *usecase) GetRelationship(ctx context.Context, userID, targetID int64) (Relationship, error) {
	return u.followRepo.GetRelationship(ctx, userID, targetID)
}

func (u *usecase) GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error) {
	ids, err := u.followRepo.GetFollowerIDs(ctx, userID)
	return ids, pkg.OrInternalError(err)
}

func (u *usecase) validateFollow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("cannot perform this action on yourself: %w", pkg.ErrBadRequest)
	}

	account, err := u.userFetcher.GetSummary(ctx, followeeID)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	if account == nil {
		return fmt.Errorf("user not found: %w", pkg.ErrBadRequest)
	}
	return nil
}

func GetFollowCacheKey(userID int64) string {
	return common.BuildCacheKey("follow", "followers", userID)
}
