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

func (r *repository) SearchPosts(ctx context.Context, params search.PostsParams) ([]search.Post, error) {
	// todo
	return nil, nil
}
