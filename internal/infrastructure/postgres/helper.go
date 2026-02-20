package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/pkg"
)

// ExecOne executes a query and expects exactly one row to be affected.
// It returns pkg.ErrNotFound if no rows were affected (RowsAffected == 0).
//
// CAUTION: 
// 1. Use this for strict updates/deletes where the entity MUST exist (e.g., UpdateProfile, UpdatePassword).
// 2. DO NOT use this for idempotent operations where "already done" is considered a success 
//    (e.g., Follow with ON CONFLICT DO NOTHING, or Unfollow/Logout).
// 3. For bulk updates that might affect zero rows naturally, use db.ExecContext instead.
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