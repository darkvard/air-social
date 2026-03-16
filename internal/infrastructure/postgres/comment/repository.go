package comment

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/comment"
	"air-social/internal/infrastructure/postgres"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, comment *comment.Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, parent_id, content, media)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	table := FromDomainComment(comment)
	if err := r.db.QueryRowContext(ctx,
		query,
		table.PostID,
		table.UserID,
		table.ParentID,
		table.Content,
		table.Media,
	).Scan(&comment.ID, &comment.CreatedAt); err != nil {
		return pkg.MapPostgresError(err)
	}

	return nil
}

func (r *repository) Update(ctx context.Context, comment *comment.Comment) error {
	query := `
		UPDATE comments
		SET content = $1,
			media = $2,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $3 AND version = $4 AND deleted_at IS NULL
		RETURNING version, updated_at
	`

	table := FromDomainComment(comment)
	if err := r.db.QueryRowContext(ctx, query,
		table.Content,
		table.Media,
		table.ID,
		table.Version,
	).Scan(
		&comment.Version,
		&comment.UpdatedAt,
	); err != nil {
		return pkg.MapPostgresError(err)
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE comments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	return postgres.ExecOne(ctx, r.db, query, id)
}

func (r *repository) GetByID(ctx context.Context, id int64) (*comment.Comment, error) {
	query := `
		SELECT 
			c.*, 
			u.full_name AS author_full_name, 
            u.avatar AS author_avatar, 
            u.verified AS author_verified
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`
	var row CommentRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return row.ToDomain(), nil
}

func (r *repository) GetComments(ctx context.Context, postID int64, params comment.GetCursorParams) ([]comment.Comment, error) {
	return r.fetchComments(ctx, params, "c.post_id = $1 AND c.parent_id IS NULL", postID)
}

func (r *repository) GetReplies(ctx context.Context, parentID int64, params comment.GetCursorParams) ([]comment.Comment, error) {
	return r.fetchComments(ctx, params, "c.parent_id = $1", parentID)
}

func (r *repository) fetchComments(ctx context.Context, params comment.GetCursorParams, baseCondition string, baseArgs ...any) ([]comment.Comment, error) {
	var args []any
	args = append(args, baseArgs...)
	argID := len(args) + 1

	var builder strings.Builder
	builder.WriteString(`
        SELECT 
            c.*, 
            u.full_name AS author_full_name, 
            u.avatar AS author_avatar, 
            u.verified AS author_verified
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE
    `)
	builder.WriteString(" " + baseCondition)
	builder.WriteString(` AND c.deleted_at IS NULL`)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND c.id %s $%d", params.Query.GetCompareOperator(), argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	fmt.Fprintf(&builder, " ORDER BY c.id %s LIMIT $%d", params.Query.GetSortOrder(), argID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []CommentRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]comment.Comment, len(rows))
	for i, v := range rows {
		result[i] = *v.ToDomain()
	}

	return result, nil
}
