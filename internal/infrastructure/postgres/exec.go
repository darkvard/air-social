package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/pkg"
)

// args must match positional parameters ($1, $2, ...)
func ExecOne(ctx context.Context, db sqlx.ExtContext, query string, args ...any) error {
	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return pkg.MapPostgresError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return pkg.ErrNotFound
	}

	return nil
}
