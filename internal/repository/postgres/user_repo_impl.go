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
        INSERT INTO users (email, username, password_hash)
        VALUES (:email, :username, :password_hash)
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
	query := ` SELECT * FROM users WHERE email = $1 `
	var u domain.User
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &u, nil
}

func (r *UserRepoImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := ` SELECT * FROM users WHERE id = $1 `
	var u domain.User
	if err := r.db.GetContext(ctx, &u, query, id); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &u, nil
}

func (r *UserRepoImpl) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET username = :username, 
			password_hash = :password_hash, 
			full_name = :full_name,
			bio = :bio,
			avatar = :avatar,
			cover_image = :cover_image,
			location = :location,
			website = :website,
			verified = :verified, 
			verified_at = :verified_at, 
			updated_at = NOW(), 
			version = version + 1
		WHERE id = :id AND version = :version
		RETURNING updated_at, version
	`
	rows, err := r.db.NamedQueryContext(ctx, query, u)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(u)
	}

	return pkg.ErrNotFound
}

func (r *UserRepoImpl) UpdateProfileImages(ctx context.Context, userID int64, url string, imageType domain.FileType) error {
	var col string
	switch imageType {
	case domain.AvatarType:
		col = "avatar"
	case domain.CoverType:
		col = "cover_image"
	default:
		return pkg.ErrInvalidData
	}
	query := `
		UPDATE users
		SET ` + col + ` = :url
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]any{
		"url": url,
		"id":  userID,
	})

	return pkg.MapPostgresError(err)
}
