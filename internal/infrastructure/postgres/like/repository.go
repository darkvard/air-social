package like

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) InsertPostLike(ctx context.Context, postID, userID int64) (bool, error) {
	query := `
		INSERT INTO post_likes (post_id, user_id) 
		VALUES ($1, $2) 
		ON CONFLICT (post_id, user_id) DO NOTHING
	`
	res, err := r.db.ExecContext(ctx, query, postID, userID)
	if err != nil {
		return false, pkg.MapPostgresError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, pkg.MapPostgresError(err)
	}

	return affected == 1, nil
}

func (r *repository) DeletePostLike(ctx context.Context, postID, userID int64) error {
	query := `
		DELETE FROM post_likes 
		WHERE post_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, postID, userID)
	return err
}

func (r *repository) InsertCommentLike(ctx context.Context, commentID, userID int64) (bool, error) {
	query := `
		INSERT INTO comment_likes (comment_id, user_id) 
		VALUES ($1, $2) 
		ON CONFLICT (comment_id, user_id) DO NOTHING
	`
	res, err := r.db.ExecContext(ctx, query, commentID, userID)
	if err != nil {
		return false, pkg.MapPostgresError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, pkg.MapPostgresError(err)
	}

	return affected == 1, nil
}

func (r *repository) DeleteCommentLike(ctx context.Context, commentID, userID int64) error {
	query := `
		DELETE FROM comment_likes 
		WHERE comment_id = $1 AND user_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, commentID, userID)
	return pkg.MapPostgresError(err)
}
