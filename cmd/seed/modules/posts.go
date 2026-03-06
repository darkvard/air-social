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
	"air-social/internal/domain/post"
)

func SeedPosts(db *sqlx.DB, userIDs []int64, cfg config.SeedConfig) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	postStmt := preparePostStmt(tx)
	mediaStmt := preparePostMediaStmt(tx)

	totalPosts := len(userIDs) * cfg.Posts.PerUser
	postIDs := make([]int64, 0, totalPosts)

	for _, userID := range userIDs {
		for range cfg.Posts.PerUser {
			var postID int64
			visibility := gofakeit.RandomString([]string{"public", "followers", "private"})
			err := postStmt.QueryRow(userID, gofakeit.Sentence(15), visibility).Scan(&postID)
			if err != nil {
				log.Panicf("insert post for user %d failed: %v", userID, err)
			}
			postIDs = append(postIDs, postID)

			// Seed media for this post
			numMedia := rand.Intn(cfg.Posts.MediaPerPost + 1)
			for range numMedia {
				mediaKey := fmt.Sprintf("posts/%d/feed_image/%s.jpg", postID, ulid.Make())
				metadata := post.MediaMetadata{
					Width:    1920,
					Height:   1080,
					Size:     int64(gofakeit.Number(100000, 5000000)),
					FileName: gofakeit.UUID() + ".jpg",
				}
				metadataJSON, _ := json.Marshal(metadata)

				_, err := mediaStmt.Exec(postID, mediaKey, "image", metadataJSON)
				if err != nil {
					log.Panicf("insert post_media for post %d failed: %v", postID, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit posts failed: %v", err)
	}

	log.Printf("Seeded: %d Posts (with media)", len(postIDs))
	return postIDs
}

func TruncatePosts(db *sqlx.DB) {
	_, err := db.Exec(`TRUNCATE TABLE posts RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Panicf("cannot clean posts data: %v", err)
	}
}

func preparePostStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO posts (user_id, content, visibility)
        VALUES ($1, $2, $3)
        RETURNING id
    `)
	if err != nil {
		log.Panicf("prepare post stmt failed: %v", err)
	}
	return stmt
}

func preparePostMediaStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO post_media (post_id, media_key, media_type, metadata)
        VALUES ($1, $2, $3, $4)
    `)
	if err != nil {
		log.Panicf("prepare post_media stmt failed: %v", err)
	}
	return stmt
}
