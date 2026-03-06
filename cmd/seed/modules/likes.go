package modules

import (
	"log"
	"math/rand"

	"air-social/cmd/seed/config"

	"github.com/jmoiron/sqlx"
)

func SeedLikes(db *sqlx.DB, postIDs, commentIDs, userIDs []int64, cfg config.SeedConfig) {
	tx := db.MustBegin()
	defer tx.Rollback()

	postLikeStmt := preparePostLikeStmt(tx)
	commentLikeStmt := prepareCommentLikeStmt(tx)

	// Seed Post Likes
	totalPostLikes := 0
	for _, postID := range postIDs {
		rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })

		numLikes := rand.Intn(cfg.Likes.PerPost + 1)
		for i := 0; i < numLikes && i < len(userIDs); i++ {
			_, err := postLikeStmt.Exec(postID, userIDs[i])
			if err != nil {
				log.Panicf("insert post_like for post %d failed: %v", postID, err)
			}
			totalPostLikes++
		}
	}

	// Seed Comment Likes
	totalCommentLikes := 0
	for _, commentID := range commentIDs {
		rand.Shuffle(len(userIDs), func(i, j int) { userIDs[i], userIDs[j] = userIDs[j], userIDs[i] })

		numLikes := rand.Intn(cfg.Likes.PerComment + 1)
		for i := 0; i < numLikes && i < len(userIDs); i++ {
			_, err := commentLikeStmt.Exec(commentID, userIDs[i])
			if err != nil {
				log.Panicf("insert comment_like for comment %d failed: %v", commentID, err)
			}
			totalCommentLikes++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit likes failed: %v", err)
	}

	log.Printf("Seeded: %d Post Likes and %d Comment Likes", totalPostLikes, totalCommentLikes)
}

func preparePostLikeStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO post_likes (post_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (post_id, user_id) DO NOTHING
    `)
	if err != nil {
		log.Panicf("prepare post_likes stmt failed: %v", err)
	}
	return stmt
}

func prepareCommentLikeStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO comment_likes (comment_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT (comment_id, user_id) DO NOTHING
    `)
	if err != nil {
		log.Panicf("prepare comment_likes stmt failed: %v", err)
	}
	return stmt
}
