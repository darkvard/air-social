package token

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/auth/token"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *token.RefreshToken) error {
	var table Table
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	args := []any{token.UserID, token.TokenHash, token.ExpiresAt, token.DeviceID}

	if err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&table); err != nil {
		return pkg.MapPostgresError(err)
	}
	*token = *table.ToDomain()

	return nil
}

func (r *repository) GetByHash(ctx context.Context, hash string) (*token.RefreshToken, error) {
	var table Table
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1`
	if err := r.db.GetContext(ctx, &table, query, hash); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return table.ToDomain(), nil
}

func (r *repository) UpdateRevoked(ctx context.Context, tokenID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), tokenID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) UpdateRevokedByUser(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), userID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND device_id = $3`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), userID, deviceID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}
