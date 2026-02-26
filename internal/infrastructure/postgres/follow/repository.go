package follow

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/follow"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, followerID int64, followeeID int64) error {
	query := `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, query, followerID, followeeID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, followerID int64, followeeID int64) error {
	query := `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`
	if _, err := r.db.ExecContext(ctx, query, followerID, followeeID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) CountFollowings(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, "follower_id", userID)
}

func (r *repository) CountFollowers(ctx context.Context, userID int64) (int64, error) {
	return r.count(ctx, "followee_id", userID)
}

func (r *repository) GetFollowers(ctx context.Context, params follow.GetFollowsParams) ([]follow.FollowUser, int64, error) {
	// f.followee_id = $2: get the people who are following Target
	// u.id = f.follower_id: join to get the information of those people
	query := fmt.Sprintf(`
		SELECT 
			u.id, u.full_name, u.avatar, u.verified,
			EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = u.id) as is_following,
			EXISTS (SELECT 1 FROM follows WHERE follower_id = u.id AND followee_id = $1) as is_follower,
			COUNT(*) OVER() as total_count
		FROM users u
		JOIN follows f ON u.id = f.follower_id
		WHERE f.followee_id = $2
		ORDER BY %s
		LIMIT $3 OFFSET $4
	`, buildSortClause(params.Paging.Sort, "f"))
	return r.fetchFollows(ctx, query, params)
}

func (r *repository) GetFollowings(ctx context.Context, params follow.GetFollowsParams) ([]follow.FollowUser, int64, error) {
	// f.follower_id = $2: get the people that Target is following
	// u.id = f.followee_id: Join to get the information of those people
	query := fmt.Sprintf(`
		SELECT 
			u.id, u.full_name, u.avatar, u.verified,
			EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = u.id) as is_following,
			EXISTS (SELECT 1 FROM follows WHERE follower_id = u.id AND followee_id = $1) as is_follower,
			COUNT(*) OVER() as total_count
		FROM users u
		JOIN follows f ON u.id = f.followee_id
		WHERE f.follower_id = $2
		ORDER BY %s
		LIMIT $3 OFFSET $4
	`, buildSortClause(params.Paging.Sort, "f"))
	return r.fetchFollows(ctx, query, params)
}

func (r *repository) GetRelationship(ctx context.Context, userID int64, targetID int64) (follow.Relationship, error) {
	query := `
		SELECT 
			EXISTS (SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2) as is_following,
			EXISTS (SELECT 1 FROM follows WHERE follower_id = $2 AND followee_id = $1) as is_follower
	`
	var res follow.Relationship
	err := r.db.GetContext(ctx, &res, query, userID, targetID)
	return res, err
}

func (r *repository) GetRelationships(ctx context.Context, userID int64, targetIDs []int64) (map[int64]follow.Relationship, error) {
	if len(targetIDs) == 0 {
		return make(map[int64]follow.Relationship), nil
	}

	query, args, err := sqlx.In(`
        SELECT follower_id, followee_id
        FROM follows
        WHERE 
            (follower_id = ? AND followee_id IN (?)) OR 
            (followee_id = ? AND follower_id IN (?))
        `, userID, targetIDs, userID, targetIDs,
	)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var follows []Table
	if err := r.db.SelectContext(ctx, &follows, query, args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make(map[int64]follow.Relationship, len(targetIDs))
	for _, p := range follows {
		var targetID int64
		var isFollowing bool
		var isFollower bool

		// determine the relationship direction relative to the current userID
		if p.FollowerID == userID {
			// case: i am following them (outgoing relationship)
			targetID = p.FolloweeID
			isFollowing = true
		} else {
			// case: they are following me (incoming relationship)
			targetID = p.FollowerID
			isFollower = true
		}

		// retrieve existing relationship status from map (default to zero-value if not present)
		res := result[targetID]

		// update flags only if true, preserving existing state for mutual follows
		// do not overwrite with false as it would wipe out previously found directions
		if isFollowing {
			res.IsFollowing = true
		}
		if isFollower {
			res.IsFollower = true
		}

		// store the updated struct back into the map
		result[targetID] = res
	}

	return result, nil
}

func (r *repository) fetchFollows(ctx context.Context, query string, params follow.GetFollowsParams) ([]follow.FollowUser, int64, error) {
	type row struct {
		ID          int64  `db:"id"`
		FullName    string `db:"full_name"`
		Avatar      string `db:"avatar"`
		IsVerified  bool   `db:"verified"`
		IsFollowing bool   `db:"is_following"`
		IsFollower  bool   `db:"is_follower"`
		TotalCount  int64  `db:"total_count"`
	}

	var rows []row
	err := r.db.SelectContext(ctx, &rows, query,
		params.ViewerID,
		params.TargetID,
		params.Paging.Limit,
		params.Paging.GetOffset(),
	)
	if err != nil {
		return nil, 0, err
	}

	if len(rows) == 0 {
		return []follow.FollowUser{}, 0, nil
	}

	total := rows[0].TotalCount
	users := make([]follow.FollowUser, len(rows))
	for i, v := range rows {
		users[i] = follow.FollowUser{
			ID:         v.ID,
			FullName:   v.FullName,
			Avatar:     v.Avatar,
			IsVerified: v.IsVerified,
			Relationship: follow.Relationship{
				IsFollowing: v.IsFollowing,
				IsFollower:  v.IsFollower,
			},
		}
	}

	return users, total, nil
}

func (r *repository) count(ctx context.Context, column string, userID int64) (int64, error) {
	var query string
	switch column {
	case "follower_id", "followee_id":
		query = fmt.Sprintf("SELECT COUNT(*) FROM follows WHERE %s = $1", column)
	default:
		return 0, fmt.Errorf("invalid column for count: %s", column)
	}
	var total int64
	err := r.db.GetContext(ctx, &total, query, userID)
	return total, err
}

func buildSortClause(sort string, tableAlias string) string {
	switch sort {
	case follow.SortOldest:
		return fmt.Sprintf("%s.created_at ASC", tableAlias)
	case follow.SortNameASC:
		return "u.username ASC"
	case follow.SortNameDESC:
		return "u.username DESC"
	default:
		return fmt.Sprintf("%s.created_at DESC", tableAlias)
	}
}
