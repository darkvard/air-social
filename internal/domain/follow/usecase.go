package follow

import (
	"context"
	"fmt"

	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type UseCase interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowings(ctx context.Context, params GetFollowsParams) (shared.OffsetPaginatedResult[FollowUser], error)
	GetFollowers(ctx context.Context, params GetFollowsParams) (shared.OffsetPaginatedResult[FollowUser], error)
}

type UserFetcher interface {
	GetSummary(ctx context.Context, id int64) (*user.UserSummary, error)
}

type Deps struct {
	FollowRepo  Repository
	UserFetcher UserFetcher
}

type usecase struct {
	followRepo  Repository
	userFetcher UserFetcher
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		followRepo:  deps.FollowRepo,
		userFetcher: deps.UserFetcher,
	}
}

func (u *usecase) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	if err := u.validateFollow(ctx, followerID, followeeID); err != nil {
		return err
	}
	return pkg.OrInternalError(u.followRepo.Create(ctx, followerID, followeeID))
}

func (u *usecase) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	if err := u.validateFollow(ctx, followerID, followeeID); err != nil {
		return err
	}
	return pkg.OrInternalError(u.followRepo.Delete(ctx, followerID, followeeID))
}

func (u *usecase) GetFollowings(ctx context.Context, params GetFollowsParams) (shared.OffsetPaginatedResult[FollowUser], error) {
	params.Paging.NormalizePagination()

	data, total, err := u.followRepo.GetFollowings(ctx, params)
	if err != nil {
		return shared.OffsetPaginatedResult[FollowUser]{}, err
	}

	return shared.NewOffsetPaginatedResult(data, total, params.Paging.Page, params.Paging.Limit), nil
}

func (u *usecase) GetFollowers(ctx context.Context, params GetFollowsParams) (shared.OffsetPaginatedResult[FollowUser], error) {
	params.Paging.NormalizePagination()

	data, total, err := u.followRepo.GetFollowers(ctx, params)
	if err != nil {
		return shared.OffsetPaginatedResult[FollowUser]{}, err
	}

	return shared.NewOffsetPaginatedResult(data, total, params.Paging.Page, params.Paging.Limit), nil
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
