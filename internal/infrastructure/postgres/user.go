package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/pkg"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (email, username, password_hash)
        VALUES (:email, :username, :password_hash)
        RETURNING id, created_at, updated_at, version
    `
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(user)
	}

	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := ` SELECT * FROM users WHERE email = $1 `
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := ` SELECT * FROM users WHERE id = $1 `
	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = pkg.TimeNowUTC()

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
			updated_at = ":updated_at", 
			version = version + 1
		WHERE id = :id AND version = :version
		RETURNING version
	`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	defer rows.Close()

	if rows.Next() {
		var newVersion int
        if err := rows.Scan(&newVersion); err != nil {
             return err
        }
        user.Version = newVersion
        return nil
	}

	return pkg.ErrNotFound
}

func (r *userRepository) UpdateProfileImages(ctx context.Context, userID int64, url string, feature domain.UploadFeature) error {
	var col string
	switch feature {
	case domain.FeatureAvatar:
		col = "avatar"
	case domain.FeatureCover:
		col = "cover_image"
	default:
		return fmt.Errorf("update profile: unsupported feature %s", feature)
	}

	query := fmt.Sprintf(`UPDATE users SET %s = :url WHERE id = :id`, col)

	_, err := r.db.NamedExecContext(ctx, query, map[string]any{
		"url": url,
		"id":  userID,
	})

	return pkg.MapPostgresError(err)
}
