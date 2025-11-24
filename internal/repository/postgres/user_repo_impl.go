package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/user"
	"air-social/pkg"
)

type UserRepoImpl struct {
	db *sqlx.DB
}

func NewUserRepoImpl(db *sqlx.DB) *UserRepoImpl {
	return &UserRepoImpl{db: db}
}

func (r *UserRepoImpl) Create(ctx context.Context, user *user.User) error {
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

func (r *UserRepoImpl) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, username, password_hash, profile, created_at, updated_at, version
		FROM users
		WHERE email = $1
		LIMIT 1
	`
	var u user.User
	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pkg.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepoImpl) GetByID(ctx context.Context, id int64) (*user.User, error) {
	return nil, nil
}

func (r *UserRepoImpl) Update(ctx context.Context, u *user.User) error {
	return nil
}
