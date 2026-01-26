package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/pkg"
)

type tokenRepository struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) *tokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, t domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_id)
		VALUES (:user_id, :token_hash, :expires_at, :device_id)
	`
	if _, err := r.db.NamedExecContext(ctx, query, t); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *tokenRepository) GetByHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, device_id
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var token domain.RefreshToken
	if err := r.db.GetContext(ctx, &token, query, hash); err != nil {
		return domain.RefreshToken{}, pkg.MapPostgresError(err)
	}
	return token, nil
}

func (r *tokenRepository) UpdateRevoked(ctx context.Context, id int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), id); err != nil {
		return err
	}
	return nil
}

func (r *tokenRepository) UpdateRevokedByUser(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), userID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *tokenRepository) UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND device_id = $3`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), userID, deviceID); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *tokenRepository) DeleteExpiredAndRevoked(ctx context.Context, expiredBefore time.Time, revokedBefore time.Time) error {
	query := `
        DELETE FROM refresh_tokens 
        WHERE (revoked_at < $1) OR (expires_at < $2)
    `
	if _, err := r.db.ExecContext(ctx, query, revokedBefore, expiredBefore); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}
