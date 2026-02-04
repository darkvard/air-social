package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
)

type followRepository struct {
	db *sqlx.DB
}

func NewFollowRepository(db *sqlx.DB) *followRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Create(ctx context.Context, follow *domain.Follow) error {
	query := `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING
		RETURNING created_at
	`

	err := r.db.GetContext(ctx, &follow.CreatedAt, query, follow.FollowerID, follow.FolloweeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // conflict (exists)
			return nil
		}
		return err
	}

	return nil
}

func (r *followRepository) Delete(ctx context.Context, followerID int64, followeeID int64) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2	
	`
	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	return err
}

func (r *followRepository) IsFollowing(ctx context.Context, followerID int64, followeeID int64) (bool, error) {
	return true, nil
}

func (r *followRepository) GetFollowings(ctx context.Context, userID int64, limit int, offset int) ([]domain.User, error) {
	return []domain.User{}, nil
}

func (r *followRepository) GetFollowers(ctx context.Context, userID int64, limit int, offset int) ([]domain.User, error) {
	return []domain.User{}, nil
}

func (r *followRepository) CountFollowings(ctx context.Context, userID int64) (int64, error) {
	return 1, nil
}

func (r *followRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return 1, nil
}
