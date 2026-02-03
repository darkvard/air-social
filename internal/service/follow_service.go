package service

import (
	"context"

	"air-social/internal/domain"
)

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowings(ctx context.Context, userID int64, page, limit int) ([]domain.User, error)
	GetFollowers(ctx context.Context, userID int64, page, limit int) ([]domain.User, error)
}

type FollowServiceImpl struct {
	followRepo domain.FollowRepository
}

func NewFollowService(followRepo domain.FollowRepository) *FollowServiceImpl {
	return &FollowServiceImpl{
		followRepo: followRepo,
	}
}


func (s *FollowServiceImpl) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	panic("not implemented") // TODO: Implement
}

func (s *FollowServiceImpl) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	panic("not implemented") // TODO: Implement
}

func (s *FollowServiceImpl) GetFollowings(ctx context.Context, userID int64, page int, limit int) ([]domain.User, error) {
	panic("not implemented") // TODO: Implement
}

func (s *FollowServiceImpl) GetFollowers(ctx context.Context, userID int64, page int, limit int) ([]domain.User, error) {
	panic("not implemented") // TODO: Implement
}

