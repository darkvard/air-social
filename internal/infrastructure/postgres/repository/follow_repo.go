package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/postgres/model"
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Conflict means relation exists, treat as success
		}
		return err
	}
	return nil
}

func (r *followRepository) Delete(ctx context.Context, followerID, followeeID int64) error {
	query := `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`
	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	return err
}

func (r *followRepository) GetFollowers(ctx context.Context, params domain.FollowParams) ([]domain.User, error) {
	return r.fetchUsers(ctx, params, "follower_id", "followee_id")
}

func (r *followRepository) GetFollowings(ctx context.Context, params domain.FollowParams) ([]domain.User, error) {
	return r.fetchUsers(ctx, params, "followee_id", "follower_id")
}

func (r *followRepository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, userID, "followee_id")
}

func (r *followRepository) CountFollowings(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, userID, "follower_id")
}

func (r *followRepository) IsFollowing(ctx context.Context, userID int64, targetIDs []int64) (map[int64]bool, error) {
	query := `SELECT followee_id FROM follows WHERE follower_id = ? AND followee_id IN (?)`
	return r.checkRelationship(ctx, query, userID, targetIDs)
}

func (r *followRepository) IsFollowedBy(ctx context.Context, userID int64, targetIDs []int64) (map[int64]bool, error) {
	query := `SELECT follower_id FROM follows WHERE followee_id = ? AND follower_id IN (?)`
	return r.checkRelationship(ctx, query, userID, targetIDs)
}

// Internal helpers

// fetchUsers executes a join query to retrieve users with sorting and pagination.
// joinCol: the column in 'follows' table to join with 'users' table.
// whereCol: the column in 'follows' table to filter by target ID.
func (r *followRepository) fetchUsers(ctx context.Context, params domain.FollowParams, joinCol, whereCol string) ([]domain.User, error) {
	orderBy := r.buildSortClause(params.Sort, "f")

	query := fmt.Sprintf(`
		SELECT u.*
		FROM users u
		JOIN follows f ON u.id = f.%s
		WHERE f.%s = $1
		ORDER BY %s
		LIMIT $2 OFFSET $3
	`, joinCol, whereCol, orderBy)

	args := []any{params.TargetUserID, params.Limit, params.GetOffset()}

	var dbUsers []model.User
	if err := r.db.SelectContext(ctx, &dbUsers, query, args...); err != nil {
		return nil, err
	}

	return model.MapToDomainUsers(dbUsers), nil
}

func (r *followRepository) buildSortClause(sort string, tableAlias string) string {
	switch sort {
	case domain.SortOldest:
		return fmt.Sprintf("%s.created_at ASC", tableAlias)
	case domain.SortNameASC:
		return "u.username ASC"
	case domain.SortNameDESC:
		return "u.username DESC"
	default:
		return fmt.Sprintf("%s.created_at DESC", tableAlias)
	}
}

func (r *followRepository) count(ctx context.Context, userID int64, whereCol string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM follows WHERE %s = $1", whereCol)
	var total int64
	err := r.db.GetContext(ctx, &total, query, userID)
	return total, err
}

func (r *followRepository) checkRelationship(ctx context.Context, rawQuery string, userID int64, targetIDs []int64) (map[int64]bool, error) {
	if len(targetIDs) == 0 {
		return make(map[int64]bool), nil
	}

	query, args, err := sqlx.In(rawQuery, userID, targetIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query) // map placeholder: (?) -> ($index)

	var matchedIDs []int64
	if err := r.db.SelectContext(ctx, &matchedIDs, query, args...); err != nil {
		return nil, err
	}

	result := make(map[int64]bool)
	for _, id := range matchedIDs {
		result[id] = true
	}
	return result, nil
}
