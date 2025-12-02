package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserRepoImpl struct {
	db *sqlx.DB
}

func NewUserRepoImpl(db *sqlx.DB) *UserRepoImpl {
	return &UserRepoImpl{db: db}
}

func (r *UserRepoImpl) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (email, username, password_hash, profile)
        VALUES (:email, :username, :password_hash, :profile)
        RETURNING id, created_at, updated_at, version
    `
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(user)
	}

	return pkg.ErrDatabase
}

func (r *UserRepoImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, profile, created_at, updated_at, version
		FROM users
		WHERE email = $1
	`
	var u domain.User
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
		pkg.Log().Infow("repo", "repo", u)

	return &u, nil
}

func (r *UserRepoImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, nil
}

func (r *UserRepoImpl) Update(ctx context.Context, u *domain.User) error {
	return nil
}
