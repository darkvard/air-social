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

func (r *postRepository) GetByID(ctx context.Context, postID, userID int64) (*domain.Post, error) {
	// todo: post view by user id (get isLiked - Open-Closed Principle)
	post, err := r.getPost(ctx, postID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	media, err := r.getPostMedia(ctx, postID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	result := post.ToDomain()
	if len(media) > 0 {
		result.Media = make([]domain.PostMedia, len(media))
		for i, v := range media {
			result.Media[i] = *v.ToDomain()
		}
	}

	return result, nil
}

func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	query := `
		UPDATE posts 
		SET content = $1, visibility = $2, updated_at = NOW(), version = version + 1
		WHERE id = $3 AND version = $4 AND deleted_at IS NULL
		RETURNING updated_at, version
	`

	args := []any{post.Content, post.Visibility, post.ID, post.Version}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&post.UpdatedAt, &post.Version)
	if err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *postRepository) IsOwner(ctx context.Context, postID, userID int64) (bool, error) {
	var ownerID int64
	query := `SELECT user_id FROM posts WHERE id = $1 AND deleted_at IS NULL`

	if err := r.db.GetContext(ctx, &ownerID, query, postID); err != nil {
		return false, pkg.MapPostgresError(err)
	}

	return ownerID == userID, nil
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

func (r *postRepository) GetUserPosts(ctx context.Context, userID int64, params domain.CursorQueryParams) ([]domain.Post, error) {
	var posts []model.Post
	fetchLimit := params.Limit + 1 // +1 for HasNextPage

	var query string
	var args []any
	if params.Cursor > 0 {
		query = `
			SELECT * FROM posts 
			WHERE user_id = $1 AND id < $2 AND deleted_at IS NULL 
			ORDER BY id DESC 
			LIMIT $3`
		args = []any{userID, params.Cursor, fetchLimit}
	} else {
		query = `
			SELECT * FROM posts 
			WHERE user_id = $1 AND deleted_at IS NULL
			ORDER BY id DESC 
			LIMIT $2`
		args = []any{userID, fetchLimit}
	}

	if err := r.db.SelectContext(ctx, &posts, query, args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	if len(posts) == 0 {
		return []domain.Post{}, nil
	}

	mediaMap, err := r.getPostMediaMap(ctx, posts)
	if err != nil {
		return nil, err
	}
	return r.toDomainList(posts, mediaMap), nil
}

// Internal helper

func (r *postRepository) getPostMediaMap(ctx context.Context, posts []model.Post) (map[int64][]domain.PostMedia, error) {
	// ids
	ids := make([]int64, len(posts))
	for i := range posts {
		ids[i] = posts[i].ID
	}

	// query
	query, args, err := sqlx.In(`SELECT * FROM post_media WHERE post_id IN (?)`, ids)
	if err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	query = r.db.Rebind(query) // (?) -> ($index)

	var medias []model.PostMedia
	if err := r.db.SelectContext(ctx, &medias, query, args...); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	// map
	mediaMap := make(map[int64][]domain.PostMedia, len(posts))
	for i := range medias {
		m := medias[i]
		mediaMap[m.PostID] = append(mediaMap[m.PostID], *m.ToDomain())
	}
	return mediaMap, nil
}

func (r *postRepository) toDomainList(posts []model.Post, mediaMap map[int64][]domain.PostMedia) []domain.Post {
	result := make([]domain.Post, len(posts))
	for i := range posts {
		domainPost := posts[i].ToDomain()
		domainPost.Media = mediaMap[posts[i].ID]
		result[i] = *domainPost
	}
	return result
}

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

func (r *postRepository) getPost(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	query := `SELECT * FROM posts WHERE id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &post, query, id); err != nil {
		return nil, pkg.MapPostgresError(err)
	}
	return &post, nil
}

func (r *postRepository) getPostMedia(ctx context.Context, postID int64) ([]model.PostMedia, error) {
	var media []model.PostMedia
	query := `SELECT * FROM post_media WHERE post_id = $1 ORDER BY id ASC`

	if err := r.db.SelectContext(ctx, &media, query, postID); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	return media, nil
}
