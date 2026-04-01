package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/search"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) SearchUsers(ctx context.Context, params search.UsersParams) ([]search.User, error) {
	var args []any
	var builder strings.Builder
	args = append(args, "%"+params.Search+"%")
	argsID := 2

	builder.WriteString(`SELECT id, username, full_name, avatar, verified FROM users WHERE (username ILIKE $1 OR full_name ILIKE $1)`)
	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND id %s $%d", params.Query.GetCompareOperator(), argsID)
		args = append(args, params.Query.Cursor)
		argsID++
	}

	fmt.Fprintf(&builder, " ORDER BY id %s LIMIT $%d", params.Query.GetSortOrder(), argsID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []userRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]search.User, len(rows))
	for i, v := range rows {
		result[i] = v.ToDomain()
	}

	return result, nil
}

func (r *repository) SearchPostIDs(ctx context.Context, params search.PostsParams) ([]int64, error) {
	var args []any
	var builder strings.Builder

	// $1 = search query, $2 = viewerID (for followers visibility check)
	args = append(args, params.Search, params.ViewerID)
	argsID := 3

	builder.WriteString(`
		SELECT p.id
		FROM posts p
		WHERE to_tsvector('simple', p.content) @@ plainto_tsquery('simple', $1)
		  AND p.deleted_at IS NULL
		  AND (
		    p.visibility = 'public'
		    OR (
		      p.visibility = 'followers'
		      AND EXISTS (
		        SELECT 1 FROM follows
		        WHERE follower_id = $2 AND followee_id = p.user_id
		      )
		    )
		  )`)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND p.id %s $%d", params.Query.GetCompareOperator(), argsID)
		args = append(args, params.Query.Cursor)
		argsID++
	}

	fmt.Fprintf(&builder, " ORDER BY p.id %s LIMIT $%d", params.Query.GetSortOrder(), argsID)
	args = append(args, params.Query.GetFetchLimit())

	var ids []int64
	if err := r.db.SelectContext(ctx, &ids, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	return ids, nil
}
