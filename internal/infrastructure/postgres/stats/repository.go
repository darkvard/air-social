package stats

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/stats"
	"air-social/pkg"
)

// repository handles statistics persistence.
//
// NOTE:
//
// This repository uses a PostgreSQL bulk UPSERT pattern:
//
//	Go slices → PostgreSQL arrays → UNNEST → rows → INSERT ... ON CONFLICT
//
// UNNEST converts arrays into a temporary row set so we can batch
// insert/update many records in a single query.
//
// This pattern is highly efficient for high-throughput counter aggregation
// (likes, comments, shares) because it minimizes database round-trips.
type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *repository {
	return &repository{db: db}
}

func (r *repository) BulkUpsertPostStats(ctx context.Context, params stats.PostParams) error {
	if len(params.IDs) == 0 {
		return nil
	}

	// Defensive validation to prevent malformed batch writes
	if len(params.IDs) != len(params.Likes) ||
		len(params.IDs) != len(params.Comments) ||
		len(params.IDs) != len(params.Shares) {
		return errors.New("post stats arrays length mismatch")
	}

	query := `
		INSERT INTO post_stats (post_id, likes_count, comments_count, shares_count)
		SELECT id, l, c, s
		FROM UNNEST($1::bigint[], $2::int[], $3::int[], $4::int[]) AS t(id, l, c, s)

		ON CONFLICT (post_id) DO UPDATE
		SET
			likes_count = GREATEST(0, post_stats.likes_count + EXCLUDED.likes_count),
			comments_count = GREATEST(0, post_stats.comments_count + EXCLUDED.comments_count),
			shares_count = GREATEST(0, post_stats.shares_count + EXCLUDED.shares_count),
			updated_at = NOW()
		WHERE
			EXCLUDED.likes_count != 0
			OR EXCLUDED.comments_count != 0
			OR EXCLUDED.shares_count != 0;
	`

	_, err := r.db.ExecContext(ctx, query,
		params.IDs,
		params.Likes,
		params.Comments,
		params.Shares,
	)

	return pkg.MapPostgresError(err)
}

func (r *repository) BulkUpsertCommentStats(ctx context.Context, params stats.CommentParams) error {
	if len(params.IDs) == 0 {
		return nil
	}

	// Defensive validation
	if len(params.IDs) != len(params.Likes) ||
		len(params.IDs) != len(params.Replies) {
		return errors.New("comment stats arrays length mismatch")
	}

	query := `
		INSERT INTO comment_stats (comment_id, likes_count, replies_count)
		SELECT id, l, r
		FROM UNNEST($1::bigint[], $2::int[], $3::int[]) AS t(id, l, r)

		ON CONFLICT (comment_id) DO UPDATE
		SET
			likes_count = GREATEST(0, comment_stats.likes_count + EXCLUDED.likes_count),
			replies_count = GREATEST(0, comment_stats.replies_count + EXCLUDED.replies_count),
			updated_at = NOW()
		WHERE
			EXCLUDED.likes_count != 0
			OR EXCLUDED.replies_count != 0;
	`

	_, err := r.db.ExecContext(ctx, query,
		params.IDs,
		params.Likes,
		params.Replies,
	)

	return pkg.MapPostgresError(err)
}

func (r *repository) GetPostsStats(ctx context.Context, postIDs []int64) ([]stats.PostStats, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	query := `SELECT * FROM post_stats WHERE post_id = ANY($1::bigint[])`

	var rows []PostTable
	if err := r.db.SelectContext(ctx, &rows, query, postIDs); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]stats.PostStats, len(rows))
	for i, row := range rows {
		result[i] = *row.ToDomain()
	}
	return result, nil
}

func (r *repository) GetCommentsStats(ctx context.Context, commentIDs []int64) ([]stats.CommentStats, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}

	query := `SELECT * FROM comment_stats WHERE comment_id = ANY($1::bigint[])`

	var rows []CommentTable
	if err := r.db.SelectContext(ctx, &rows, query, commentIDs); err != nil {
		return nil, pkg.MapPostgresError(err)
	}

	result := make([]stats.CommentStats, len(rows))
	for i, row := range rows {
		result[i] = *row.ToDomain()
	}
	return result, nil
}
