package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/pkg"
)

type TokenRepoImpl struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) *TokenRepoImpl {
	return &TokenRepoImpl{db: db}
}

func (r *TokenRepoImpl) Create(ctx context.Context, t *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, device_id)
		VALUES (:user_id, :token_hash, :expires_at, :device_id)
	`
	if _, err := r.db.NamedExecContext(ctx, query, t); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *TokenRepoImpl) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	var t domain.RefreshToken
	if err := r.db.GetContext(ctx, &t, query, hash); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &t, nil
}

func (r *TokenRepoImpl) UpdateRevoked(ctx context.Context, id int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}

func (r *TokenRepoImpl) UpdateRevokedByUser(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}
	return nil
}

func (r *TokenRepoImpl) UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND device_id = $2`
	if _, err := r.db.ExecContext(ctx, query, userID, deviceID); err != nil {
		return err
	}
	return nil
}

func (r *TokenRepoImpl) DeleteExpiredAndRevoked(ctx context.Context, expiredBefore time.Time, revokedBefore time.Time) error {
    query := `
        DELETE FROM refresh_tokens 
        WHERE (revoked_at < $1) OR (expires_at < $2)
    `
    if _, err := r.db.ExecContext(ctx, query, revokedBefore, expiredBefore); err != nil {
        return err
    }
    return nil
}
