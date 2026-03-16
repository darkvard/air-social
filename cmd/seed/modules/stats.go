package modules

import (
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedStats calculates and populates the post_stats and comment_stats tables.
// This function should be run AFTER all posts, comments, likes, and shares have been seeded,
// as it calculates aggregates based on that source-of-truth data.
func SeedStats(db *sqlx.DB) {
	log.Println("Calculating and seeding aggregate stats...")

	// Use a transaction to ensure both stats tables are updated together.
	tx := db.MustBegin()
	defer tx.Rollback()

	// 1. Calculate and upsert post_stats
	// This query joins posts with aggregated counts of likes, root comments, and shares.
	postStatsQuery := `
		INSERT INTO post_stats (post_id, likes_count, comments_count, shares_count)
		SELECT 
			p.id,
			COALESCE(l.count, 0),
			COALESCE(c.count, 0),
			COALESCE(s.count, 0)
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count FROM post_likes GROUP BY post_id
		) l ON p.id = l.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as count FROM comments WHERE parent_id IS NULL GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN (
			SELECT original_post_id, COUNT(*) as count FROM posts WHERE original_post_id IS NOT NULL GROUP BY original_post_id
		) s ON p.id = s.original_post_id
		ON CONFLICT (post_id) DO UPDATE SET
			likes_count = EXCLUDED.likes_count,
			comments_count = EXCLUDED.comments_count,
			shares_count = EXCLUDED.shares_count;
	`
	if _, err := tx.Exec(postStatsQuery); err != nil {
		log.Panicf("failed to seed post_stats: %v", err)
	}

	// 2. Calculate and upsert comment_stats
	// This query joins comments with aggregated counts of likes and replies.
	commentStatsQuery := `
		INSERT INTO comment_stats (comment_id, likes_count, replies_count)
		SELECT
			c.id,
			COALESCE(l.count, 0),
			COALESCE(r.count, 0)
		FROM comments c
		LEFT JOIN (
			SELECT comment_id, COUNT(*) as count FROM comment_likes GROUP BY comment_id
		) l ON c.id = l.comment_id
		LEFT JOIN (
			SELECT parent_id, COUNT(*) as count FROM comments WHERE parent_id IS NOT NULL GROUP BY parent_id
		) r ON c.id = r.parent_id
		ON CONFLICT (comment_id) DO UPDATE SET
			likes_count = EXCLUDED.likes_count,
			replies_count = EXCLUDED.replies_count;
	`
	if _, err := tx.Exec(commentStatsQuery); err != nil {
		log.Panicf("failed to seed comment_stats: %v", err)
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit for stats failed: %v", err)
	}

	log.Println("Stats seeded successfully.")
}
