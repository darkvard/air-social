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

func (r *repository) GetPostLiked(ctx context.Context, postIDs []int64, userID int64) ([]int64, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`SELECT post_id FROM post_likes WHERE user_id = ? AND post_id IN (?)`, userID, postIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var likedIDs []int64
	err = r.db.SelectContext(ctx, &likedIDs, query, args...)
	return likedIDs, pkg.MapPostgresError(err)
}

func (r *repository) GetCommentLiked(ctx context.Context, commentIDs []int64, userID int64) ([]int64, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`SELECT comment_id FROM comment_likes WHERE user_id = ? AND comment_id IN (?)`, userID, commentIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var likedIDs []int64
	err = r.db.SelectContext(ctx, &likedIDs, query, args...)
	return likedIDs, pkg.MapPostgresError(err)
}
