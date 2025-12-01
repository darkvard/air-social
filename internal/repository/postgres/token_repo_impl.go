package postgres

import (
	"context"

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
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES (:user_id, :token_hash, :expires_at)
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

func (r *TokenRepoImpl) Revoke(ctx context.Context, id int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}

func (r *TokenRepoImpl) RevokeAllByUser(ctx context.Context, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return err
	}
	return nil
}
