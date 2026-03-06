package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"

	"air-social/cmd/seed/config"
)

func SeedComments(db *sqlx.DB, postIDs, userIDs []int64, cfg config.SeedConfig) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	commentStmt := prepareCommentStmt(tx)

	totalComments := len(postIDs) * cfg.Comments.PerPost
	commentIDs := make([]int64, 0, totalComments)

	for _, postID := range postIDs {
		// Store root comments for this post to create replies
		rootCommentsForPost := make([]int64, 0)

		for range cfg.Comments.PerPost {
			var commentID int64
			commenterID := userIDs[rand.Intn(len(userIDs))]
			content := gofakeit.Sentence(10)

			var parentID *int64
			// 20% chance to be a reply, if there are root comments available
			if len(rootCommentsForPost) > 0 && rand.Float32() < 0.2 {
				pID := rootCommentsForPost[rand.Intn(len(rootCommentsForPost))]
				parentID = &pID
			}

			// Seed media for this comment (JSON column in comments table)
			var mediaJSON []byte = []byte("[]")
			if rand.Intn(100) < cfg.Comments.MediaPerComment*20 {
				// Note: We use a placeholder ID or random string because we don't have commentID yet
				mediaKey := fmt.Sprintf("comments/seed/%s.png", ulid.Make())
				media := []map[string]string{
					{"media_key": mediaKey, "media_type": "image"},
				}
				mediaJSON, _ = json.Marshal(media)
			}

			err := commentStmt.QueryRow(postID, commenterID, content, parentID, mediaJSON).Scan(&commentID)
			if err != nil {
				log.Panicf("insert comment for post %d failed: %v", postID, err)
			}
			commentIDs = append(commentIDs, commentID)

			if parentID == nil {
				rootCommentsForPost = append(rootCommentsForPost, commentID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit comments failed: %v", err)
	}

	log.Printf("Seeded: %d Comments (with replies and media)", len(commentIDs))
	return commentIDs
}

func TruncateComments(db *sqlx.DB) {
	_, err := db.Exec(`TRUNCATE TABLE comments RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Panicf("cannot clean comments data: %v", err)
	}
}

func prepareCommentStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO comments (post_id, user_id, content, parent_id, media)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `)
	if err != nil {
		log.Panicf("prepare comment stmt failed: %v", err)
	}
	return stmt
}
