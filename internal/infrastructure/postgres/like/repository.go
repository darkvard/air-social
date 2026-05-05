package like

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/like"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) InsertPostLike(ctx context.Context, postID, userID int64) (bool, int64, error) {
	query := `
        WITH inserted_like AS (
            INSERT INTO post_likes (post_id, user_id) 
            VALUES ($1, $2) 
            ON CONFLICT (post_id, user_id) DO NOTHING
            RETURNING post_id
        )
        SELECT p.user_id as owner_id
        FROM posts p
        INNER JOIN inserted_like il ON p.id = il.post_id;
    `

	var ownerID int64
	if err := r.db.QueryRowContext(ctx, query, postID, userID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, 0, nil
		}
		return false, 0, pkg.MapPostgresError(err)
	}
	return true, ownerID, nil
}

func (r *repository) DeletePostLike(ctx context.Context, postID, userID int64) (int64, error) {
	query := `
		DELETE FROM post_likes pl
		USING posts p
		WHERE pl.post_id = $1 AND pl.user_id = $2 AND p.id = pl.post_id
		RETURNING p.user_id
	`
	var ownerID int64
	if err := r.db.QueryRowContext(ctx, query, postID, userID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, pkg.MapPostgresError(err)
	}
	return ownerID, nil
}

func (r *repository) InsertCommentLike(ctx context.Context, commentID, userID int64) (bool, int64, error) {
	query := `
		WITH inserted_like AS (
			INSERT INTO comment_likes (comment_id, user_id) 
			VALUES ($1, $2) 
			ON CONFLICT (comment_id, user_id) DO NOTHING
			RETURNING comment_id
		)
		SELECT c.user_id as owner_id
		FROM comments c
		INNER JOIN inserted_like il ON c.id = il.comment_id;
	`

	var ownerID int64
	if err := r.db.QueryRowContext(ctx, query, commentID, userID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, 0, nil
		}
		return false, 0, pkg.MapPostgresError(err)
	}
	return true, ownerID, nil
}

func (r *repository) DeleteCommentLike(ctx context.Context, commentID, userID int64) (int64, error) {
	query := `
		DELETE FROM comment_likes cl
		USING comments c
		WHERE cl.comment_id = $1 AND cl.user_id = $2 AND c.id = cl.comment_id
		RETURNING c.user_id
	`
	var ownerID int64
	if err := r.db.QueryRowContext(ctx, query, commentID, userID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, pkg.MapPostgresError(err)
	}
	return ownerID, nil
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

func (r *repository) GetPostLikers(ctx context.Context, params like.GetCursorParams) ([]like.Liker, error) {
	// JOIN: POST_ID (POST_LIKE) = USER_ID (USERS)

	var args []any
	var builder strings.Builder
	args = append(args, params.TargetID)
	argID := len(args) + 1

	builder.WriteString(`
		SELECT
			p.id AS like_id,
			p.user_id,
			u.full_name, 
			u.avatar, 
			u.verified AS is_verified
		FROM post_likes p
		JOIN users u ON p.user_id = u.id
		WHERE p.post_id = $1
	`)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND p.id %s $%d", params.Query.GetCompareOperator(), argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	fmt.Fprintf(&builder, " ORDER BY p.id %s LIMIT $%d", params.Query.GetSortOrder(), argID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []LikerRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]like.Liker, len(rows))
	for i, v := range rows {
		result[i] = *v.ToDomain()
	}

	return result, nil
}

func (r *repository) GetCommentLikers(ctx context.Context, params like.GetCursorParams) ([]like.Liker, error) {
	var args []any
	var builder strings.Builder
	args = append(args, params.TargetID)
	argID := len(args) + 1

	builder.WriteString(`
		SELECT
			c.id AS like_id,
			c.user_id,
			u.full_name, 
			u.avatar, 
			u.verified AS is_verified
		FROM comment_likes c
		JOIN users u ON c.user_id = u.id
		WHERE c.comment_id = $1
	`)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND c.id %s $%d", params.Query.GetCompareOperator(), argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	fmt.Fprintf(&builder, " ORDER BY c.id %s LIMIT $%d", params.Query.GetSortOrder(), argID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []LikerRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]like.Liker, len(rows))
	for i, v := range rows {
		result[i] = *v.ToDomain()
	}

	return result, nil
}
