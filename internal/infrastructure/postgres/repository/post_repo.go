package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/postgres/model"
	"air-social/pkg"
)

type postRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) *postRepository {
	return &postRepository{db: db}
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	return nil, nil
}

func (r *postRepository) GetByUserID(ctx context.Context, userID int64, cursor int64, limit int) ([]domain.Post, error) {
	return nil, nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (r *postRepository) IsOwner(ctx context.Context, postID int64, userID int64) (bool, error) {
	return false, nil
}

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return pkg.MapPostgresError(err)
	}

	defer tx.Rollback()

	if err := r.insertPost(ctx, tx, post); err != nil {
		return pkg.MapPostgresError(err)
	}

	if len(post.Media) > 0 {
		if err := r.insertPostMedia(ctx, tx, post.ID, post.Media); err != nil {
			return pkg.MapPostgresError(err)
		}
	}
	return tx.Commit()
}

// Internal helper

func (r *postRepository) insertPost(ctx context.Context, tx *sqlx.Tx, post *domain.Post) error {
	query := `
		INSERT INTO posts (user_id, content, visibility)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	args := []any{post.UserID, post.Content, string(post.Visibility)}

	err := tx.QueryRowContext(ctx, query, args...).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	return err
}

func (r *postRepository) insertPostMedia(ctx context.Context, tx *sqlx.Tx, postID int64, media []domain.PostMedia) error {
	mediaModels := make([]model.PostMedia, len(media))

	for i := range media {
		media[i].PostID = postID
		dbModel := model.FromDomainPostMedia(&media[i])
		mediaModels[i] = *dbModel
	}

	query := `
		INSERT INTO post_media (post_id, media_key, media_type, metadata)
		VALUES (:post_id, :media_key, :media_type, :metadata)
	`

	_, err := tx.NamedExecContext(ctx, query, mediaModels)
	return err
}
