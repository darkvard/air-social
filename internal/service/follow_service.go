package service

import (
	"context"
	"fmt"

	"air-social/internal/domain"
	"air-social/pkg"
)

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowings(ctx context.Context, userID int64, page, limit int) ([]domain.User, error)
	GetFollowers(ctx context.Context, userID int64, page, limit int) ([]domain.User, error)
}

type FollowServiceImpl struct {
	followRepo domain.FollowRepository
	userRepo   domain.UserRepository
}

func NewFollowService(followRepo domain.FollowRepository, userRepo domain.UserRepository) *FollowServiceImpl {
	return &FollowServiceImpl{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

func (s *FollowServiceImpl) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("%w: cannot follow yourself", pkg.ErrBadRequest)
	}

	user, err := s.userRepo.GetByID(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("%w: user not found", pkg.ErrBadRequest)
	}

	follow := &domain.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}

	return s.followRepo.Create(ctx, follow)
}

func (s *FollowServiceImpl) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("%w: cannot unfollow yourself", pkg.ErrBadRequest)
	}

	return s.followRepo.Delete(ctx, followerID, followeeID)
}

func (s *FollowServiceImpl) GetFollowings(ctx context.Context, userID int64, page int, limit int) ([]domain.User, error) {
	return []domain.User{}, nil
}

func (s *FollowServiceImpl) GetFollowers(ctx context.Context, userID int64, page int, limit int) ([]domain.User, error) {
	return []domain.User{}, nil
}
