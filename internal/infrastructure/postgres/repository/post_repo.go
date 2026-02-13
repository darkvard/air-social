package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
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

func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	return nil
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
