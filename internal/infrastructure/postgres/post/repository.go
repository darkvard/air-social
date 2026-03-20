package post

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/post"
	"air-social/pkg"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, post *post.Post) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	defer tx.Rollback()

	if err := r.insertPost(ctx, tx, post); err != nil {
		return pkg.MapPostgresError(err)
	}

	if len(post.Media) > 0 {
		if err := r.insertMedia(ctx, tx, post.ID, post.Media); err != nil {
			return pkg.MapPostgresError(err)
		}
	}
	return tx.Commit()
}

func (r *repository) Update(ctx context.Context, post *post.Post) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	defer tx.Rollback()

	if err := tx.QueryRowContext(ctx, `
        UPDATE posts 
        SET content = $1, visibility = $2, updated_at = $3, version = version + 1
        WHERE id = $4 AND version = $5 AND deleted_at IS NULL
        RETURNING updated_at, version
    `,
		post.Content,
		string(post.Visibility),
		pkg.TimeNowUTC(),
		post.ID,
		post.Version,
	).Scan(&post.UpdatedAt, &post.Version); err != nil {
		return pkg.MapPostgresError(err)
	}

	if post.Media != nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM post_media WHERE post_id = $1`, post.ID); err != nil {
			return pkg.MapPostgresError(err)
		}
		if err := r.insertMedia(ctx, tx, post.ID, post.Media); err != nil {
			return pkg.MapPostgresError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE posts SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	if _, err := r.db.ExecContext(ctx, query, pkg.TimeNowUTC(), id); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, postID int64) (*post.Post, error) {
	var table PostTable
	query := `SELECT * FROM posts WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &table, query, postID); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return table.ToDomain(), nil
}

func (r *repository) GetDetail(ctx context.Context, postID int64) (*post.Post, error) {
	post, err := r.getPost(ctx, postID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	media, err := r.getMedia(ctx, postID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	post.Media = media
	return post, nil
}

func (r *repository) GetUserPosts(ctx context.Context, params post.GetCursorParams) ([]post.Post, error) {
	posts, err := r.getUserPosts(ctx, params)
	if err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	if len(posts) == 0 {
		return []post.Post{}, nil
	}

	ids := make([]int64, len(posts))
	for i, v := range posts {
		ids[i] = v.ID
	}
	mediaMap, err := r.getUserPostMedia(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		if m, ok := mediaMap[posts[i].ID]; ok {
			posts[i].Media = m
		}
	}
	return posts, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []int64) ([]*post.Post, error) {
	if len(ids) == 0 {
		return []*post.Post{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT
			p.*,
			u.id AS author_id,
			u.full_name AS author_full_name,
			u.avatar AS author_avatar,
			u.verified AS author_verified
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id IN (?) AND p.deleted_at IS NULL
	`, ids)
	if err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	query = r.db.Rebind(query)

	var rows []PostDetailRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	posts := make([]*post.Post, len(rows))
	for i, v := range rows {
		posts[i] = v.ToDomain()
	}

	if len(posts) == 0 {
		return posts, nil
	}

	mediaMap, err := r.getUserPostMedia(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, p := range posts {
		if m, ok := mediaMap[p.ID]; ok {
			p.Media = m
		}
	}

	return posts, nil
}

func (r *repository) insertPost(ctx context.Context, tx *sqlx.Tx, post *post.Post) error {
	query := `
		INSERT INTO posts (user_id, content, visibility, original_post_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return tx.QueryRowContext(ctx, query,
		post.UserID,
		post.Content,
		string(post.Visibility),
		post.OriginalPostID,
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
}

func (r *repository) insertMedia(ctx context.Context, tx *sqlx.Tx, postID int64, media []post.Media) error {
	if len(media) == 0 {
		return nil
	}

	rows := make([]MediaTable, len(media))
	for i, v := range media {
		v.PostID = postID
		rows[i] = *FromDomainMedia(&v)
	}

	// NamedExec supports Bulk Insert natively.
	// When passing a slice (rows), NamedExec automatically expands this single VALUES
	// clause into multiple tuples: (?, ?, ?, ?), (?, ?, ?, ?)...
	query := `
        INSERT INTO post_media (post_id, media_key, media_type, metadata)
        VALUES (:post_id, :media_key, :media_type, :metadata)
    `

	_, err := tx.NamedExecContext(ctx, query, rows)
	return pkg.MapPostgresError(err)
}

func (r *repository) getPost(ctx context.Context, postID int64) (*post.Post, error) {
	query := `
        SELECT 
            p.*, 
            u.id AS author_id, 
            u.full_name AS author_full_name, 
            u.avatar AS author_avatar, 
            u.verified AS author_verified
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = $1 AND p.deleted_at IS NULL
    `
	var row PostDetailRow
	if err := r.db.GetContext(ctx, &row, query, postID); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return row.ToDomain(), nil
}

func (r *repository) getMedia(ctx context.Context, postID int64) ([]post.Media, error) {
	var media []MediaTable
	query := `SELECT * FROM post_media WHERE post_id = $1 ORDER BY id ASC`
	if err := r.db.SelectContext(ctx, &media, query, postID); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	result := make([]post.Media, len(media))
	for i, v := range media {
		result[i] = *v.ToDomain()
	}
	return result, nil
}

func (r *repository) getUserPosts(ctx context.Context, params post.GetCursorParams) ([]post.Post, error) {
	var args []any
	var builder strings.Builder
	args = append(args, params.UserID)
	argID := len(args) + 1

	builder.WriteString(`
        SELECT * FROM posts
        WHERE user_id = $1 AND deleted_at IS NULL
    `)

	if params.Query.Cursor > 0 {
		fmt.Fprintf(&builder, " AND id %s $%d", params.Query.GetCompareOperator(), argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	fmt.Fprintf(&builder, " ORDER BY id %s LIMIT $%d", params.Query.GetSortOrder(), argID)
	args = append(args, params.Query.GetFetchLimit())

	var posts []PostTable
	if err := r.db.SelectContext(ctx, &posts, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]post.Post, len(posts))
	for i, v := range posts {
		result[i] = *v.ToDomain()
	}

	return result, nil
}

func (r *repository) getUserPostMedia(ctx context.Context, ids []int64) (map[int64][]post.Media, error) {
	query, args, err := sqlx.In(`SELECT * FROM post_media WHERE post_id IN (?)`, ids)
	if err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	query = r.db.Rebind(query)

	var media []MediaTable
	if err := r.db.SelectContext(ctx, &media, query, args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make(map[int64][]post.Media, len(ids))
	for i := range media {
		m := media[i]
		result[m.PostID] = append(result[m.PostID], *m.ToDomain())
	}

	return result, nil
}

func (r *repository) GetPostSharers(ctx context.Context, params post.GetSharersParams) ([]post.Sharer, error) {
	var args []any
	var builder strings.Builder
	args = append(args, params.PostID)
	argID := len(args) + 1

	builder.WriteString(`
		SELECT
			p.id AS share_id,
			p.user_id,
			u.full_name,
			u.avatar,
			u.verified AS is_verified
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.original_post_id = $1 AND p.deleted_at IS NULL
	`)

	if params.Query.Cursor > 0 {
		// The cursor is the ID of the sharing post (p.id)
		fmt.Fprintf(&builder, " AND p.id %s $%d", params.Query.GetCompareOperator(), argID)
		args = append(args, params.Query.Cursor)
		argID++
	}

	// Order by the cursor column
	fmt.Fprintf(&builder, " ORDER BY p.id %s LIMIT $%d", params.Query.GetSortOrder(), argID)
	args = append(args, params.Query.GetFetchLimit())

	var rows []SharerRow
	if err := r.db.SelectContext(ctx, &rows, builder.String(), args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]post.Sharer, len(rows))
	for i := range rows {
		result[i] = *rows[i].ToDomain()
	}

	return result, nil
}
