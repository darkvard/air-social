package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
)

type followRepository struct {
	db *sqlx.DB
}

func NewFollowRepository(db *sqlx.DB) *followRepository {
	return &followRepository{db: db}
}

func (f *followRepository) Create(ctx context.Context, follow *domain.Follow) error {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) Delete(ctx context.Context, followerID int64, followeeID int64) error {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) IsFollowing(ctx context.Context, followerID int64, followeeID int64) (bool, error) {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) GetFollowings(ctx context.Context, userID int64, limit int, offset int) ([]domain.User, error) {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) GetFollowers(ctx context.Context, userID int64, limit int, offset int) ([]domain.User, error) {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) CountFollowings(ctx context.Context, userID int64) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (f *followRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	panic("not implemented") // TODO: Implement
}

