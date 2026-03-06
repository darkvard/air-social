package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"

	"air-social/cmd/seed/config"
)

const commentsBatchSize = 1000

type commentData struct {
	PostID    int64
	UserID    int64
	Content   string
	ParentID  *int64
	MediaJSON []byte
}

func SeedComments(db *sqlx.DB, postIDs, userIDs []int64, cfg config.SeedConfig) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	allCommentIDs := make([]int64, 0)

	// Helper to generate random media
	genMedia := func() []byte {
		if rand.Intn(100) < cfg.Comments.MediaPerComment*20 {
			mediaKey := fmt.Sprintf("comments/seed/%s.png", ulid.Make())
			media := []map[string]string{
				{"media_key": mediaKey, "media_type": "image"},
			}
			b, _ := json.Marshal(media)
			return b
		}
		return []byte("[]")
	}

	// --- Phase 1: Seed Root Comments ---
	var pendingRoots []commentData

	for _, postID := range postIDs {
		// Create root comments
		for i := 0; i < cfg.Comments.PerPost; i++ {
			// 80% chance to be a root comment (approx logic to separate roots/replies)
			// In batch mode, we seed ALL roots first to ensure parents exist
			if rand.Float32() < 0.8 {
				pendingRoots = append(pendingRoots, commentData{
					PostID:    postID,
					UserID:    userIDs[rand.Intn(len(userIDs))],
					Content:   gofakeit.Sentence(10),
					ParentID:  nil,
					MediaJSON: genMedia(),
				})
			}
		}
	}

	// Map PostID -> List of RootCommentIDs (to create replies later)
	rootsByPost := make(map[int64][]int64)

	// Batch Insert Roots
	for i := 0; i < len(pendingRoots); i += commentsBatchSize {
		end := i + commentsBatchSize
		if end > len(pendingRoots) {
			end = len(pendingRoots)
		}

		// We need returned ID and PostID to map them
		inserted := insertCommentsBatch(tx, pendingRoots[i:end])
		for _, c := range inserted {
			allCommentIDs = append(allCommentIDs, c.ID)
			rootsByPost[c.PostID] = append(rootsByPost[c.PostID], c.ID)
		}
	}

	// --- Phase 2: Seed Replies ---
	var pendingReplies []commentData

	for postID, rootIDs := range rootsByPost {
		if len(rootIDs) == 0 {
			continue
		}

		// 1. Random replies (the remaining 20% from PerPost config)
		repliesCount := int(float32(cfg.Comments.PerPost) * 0.2)
		for i := 0; i < repliesCount; i++ {
			parentID := rootIDs[rand.Intn(len(rootIDs))]
			pendingReplies = append(pendingReplies, commentData{
				PostID:    postID,
				UserID:    userIDs[rand.Intn(len(userIDs))],
				Content:   gofakeit.Sentence(10),
				ParentID:  &parentID,
				MediaJSON: genMedia(),
			})
		}

		// 2. Hot Thread (25 replies to one random root)
		hotParentID := rootIDs[rand.Intn(len(rootIDs))]
		for i := 0; i < 25; i++ {
			pendingReplies = append(pendingReplies, commentData{
				PostID:    postID,
				UserID:    userIDs[rand.Intn(len(userIDs))],
				Content:   gofakeit.Sentence(10),
				ParentID:  &hotParentID,
				MediaJSON: genMedia(),
			})
		}
	}

	// Batch Insert Replies
	for i := 0; i < len(pendingReplies); i += commentsBatchSize {
		end := i + commentsBatchSize
		if end > len(pendingReplies) {
			end = len(pendingReplies)
		}
		inserted := insertCommentsBatch(tx, pendingReplies[i:end])
		for _, c := range inserted {
			allCommentIDs = append(allCommentIDs, c.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit comments failed: %v", err)
	}

	log.Printf("Seeded: %d Comments (Roots & Replies)", len(allCommentIDs))
	return allCommentIDs
}

type insertedComment struct {
	ID     int64 `db:"id"`
	PostID int64 `db:"post_id"`
}

func insertCommentsBatch(tx *sqlx.Tx, comments []commentData) []insertedComment {
	if len(comments) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(comments))
	valueArgs := make([]interface{}, 0, len(comments)*5)

	for i, c := range comments {
		// $1, $2, $3, $4, $5
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, c.PostID, c.UserID, c.Content, c.ParentID, c.MediaJSON)
	}

	stmt := fmt.Sprintf("INSERT INTO comments (post_id, user_id, content, parent_id, media) VALUES %s RETURNING id, post_id", strings.Join(valueStrings, ","))

	rows, err := tx.Queryx(stmt, valueArgs...)
	if err != nil {
		log.Panicf("batch insert comments failed: %v", err)
	}
	defer rows.Close()

	var results []insertedComment
	for rows.Next() {
		var c insertedComment
		if err := rows.StructScan(&c); err != nil {
			log.Panicf("scan comment failed: %v", err)
		}
		results = append(results, c)
	}
	return results
}

func TruncateComments(db *sqlx.DB) {
	_, err := db.Exec(`TRUNCATE TABLE comments RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Panicf("cannot clean comments data: %v", err)
	}
}
