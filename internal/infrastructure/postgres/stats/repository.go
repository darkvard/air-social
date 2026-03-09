package stats

import (
	"context"

	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/stats"
	"air-social/pkg"
)

// repository handles statistics persistence.
//
// NOTE:
// This repository uses a PostgreSQL bulk UPSERT pattern:
//
//	Go slices → PostgreSQL arrays → UNNEST → rows → INSERT ... ON CONFLICT
//
// UNNEST converts arrays into a temporary row set so we can batch
// insert/update many records in a single query.
//
// This significantly reduces database round-trips and is used for
// high-throughput counter aggregation (likes, comments, shares).
type repository struct {
	db *sqlx.DB
}

func (r *repository) BulkUpsertPostStats(ctx context.Context, params stats.PostParams) error {
	if len(params.IDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO post_stats (post_id, likes_count, comments_count, shares_count)
		SELECT * FROM UNNEST($1::bigint[], $2::int[], $3::int[], $4::int[])
		ON CONFLICT (post_id) DO UPDATE
		SET 
			likes_count = post_stats.likes_count + EXCLUDED.likes_count,
			comments_count = post_stats.comments_count + EXCLUDED.comments_count,
			shares_count = post_stats.shares_count + EXCLUDED.shares_count,
			updated_at = NOW();
	`

	_, err := r.db.ExecContext(ctx, query,
		params.IDs,
		params.Likes,
		params.Comments,
		params.Shares,
	)

	if err != nil {
		return pkg.MapPostgresError(err)
	}
	return nil
}

func (r *repository) BulkUpsertCommentStats(ctx context.Context, params stats.CommentParams) error {
	if len(params.IDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO comment_stats (comment_id, likes_count, replies_count)
		SELECT * FROM UNNEST($1::bigint[], $2::int[], $3::int[])
		ON CONFLICT (comment_id) DO UPDATE
		SET 
			likes_count = comment_stats.likes_count + EXCLUDED.likes_count,
			replies_count = comment_stats.replies_count + EXCLUDED.replies_count,
			updated_at = NOW();
	`

	_, err := r.db.ExecContext(ctx, query,
		params.IDs,
		params.Likes,
		params.Replies,
	)

	return pkg.MapPostgresError(err)
}
