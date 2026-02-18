package user

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/user"
	"air-social/internal/infrastructure/postgres"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id int64) (*user.User, error) {
	var table Table
	query := `SELECT * FROM users WHERE id = $1`
	if err := r.db.GetContext(ctx, &table, query, id); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return table.ToDomain(), nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var table Table
	query := `SELECT * FROM users WHERE email = $1`
	if err := r.db.GetContext(ctx, &table, query, email); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return table.ToDomain(), nil
}

func (r *repository) Create(ctx context.Context, user *user.User) error {
	var table Table
	query := `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
	 	RETURNING id, created_at, updated_at, version
	`
	args := []any{user.Email, user.Username, user.PasswordHash}

	if err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&table); err != nil {
		return pkg.MapPostgresError(err)
	}
	*user = *table.ToDomain()

	return nil
}

func (r *repository) UpdateProfile(ctx context.Context, user *user.User) error {
	query := `
		UPDATE users
		SET username = $1,
			full_name = $2,
			bio = $3,
			location = $4,
			website = $5,
			version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version
	`

	args := []any{
		user.Username,
		user.Profile.FullName,
		user.Profile.Bio,
		user.Profile.Location,
		user.Profile.Website,
		user.ID,
		user.Version,
	}

	if err := r.db.QueryRowxContext(ctx, query, args...).Scan(&user.Version); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) UpdateAvatar(ctx context.Context, id int64, url string) error {
	query := `UPDATE users SET avatar = $1 WHERE id = $2`
	return postgres.ExecOne(ctx, r.db, query, url, id)
}

func (r *repository) UpdateCover(ctx context.Context, id int64, url string) error {
	query := `UPDATE users SET cover_image = $1 WHERE id = $2`
	return postgres.ExecOne(ctx, r.db, query, url, id)
}

func (r *repository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	return postgres.ExecOne(ctx, r.db, query, passwordHash, id)
}

func (r *repository) UpdateVerified(ctx context.Context, id int64, status bool) error {
	query := `UPDATE users SET verified = $1 WHERE id = $2`
	return postgres.ExecOne(ctx, r.db, query, status, id)
}
