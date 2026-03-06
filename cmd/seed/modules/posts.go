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
	"air-social/internal/domain/post"
)

const postsBatchSize = 1000

type postData struct {
	UserID     int64
	Content    string
	Visibility string
}

type mediaData struct {
	PostID    int64
	MediaKey  string
	MediaType string
	Metadata  []byte
}

func SeedPosts(db *sqlx.DB, userIDs []int64, cfg config.SeedConfig) []int64 {
	tx := db.MustBegin()
	defer tx.Rollback()

	totalPosts := len(userIDs) * cfg.Posts.PerUser
	allPostIDs := make([]int64, 0, totalPosts)

	// 1. Prepare and Insert Posts in Batches
	var pendingPosts []postData
	for _, userID := range userIDs {
		for i := 0; i < cfg.Posts.PerUser; i++ {
			pendingPosts = append(pendingPosts, postData{
				UserID:     userID,
				Content:    gofakeit.Sentence(15),
				Visibility: gofakeit.RandomString([]string{"public", "followers", "private"}),
			})
		}
	}

	// Batch Insert Posts
	for i := 0; i < len(pendingPosts); i += postsBatchSize {
		end := i + postsBatchSize
		if end > len(pendingPosts) {
			end = len(pendingPosts)
		}

		batchIDs := insertPostsBatch(tx, pendingPosts[i:end])
		allPostIDs = append(allPostIDs, batchIDs...)
	}

	// 2. Prepare and Insert Media in Batches based on created Post IDs
	var pendingMedia []mediaData

	for _, postID := range allPostIDs {
		numMedia := rand.Intn(cfg.Posts.MediaPerPost + 1)
		for i := 0; i < numMedia; i++ {
			mediaKey := fmt.Sprintf("posts/%d/feed_image/%s.jpg", postID, ulid.Make())
			metadata := post.MediaMetadata{
				Width:    1920,
				Height:   1080,
				Size:     int64(gofakeit.Number(100000, 5000000)),
				FileName: gofakeit.UUID() + ".jpg",
			}
			metadataJSON, _ := json.Marshal(metadata)

			pendingMedia = append(pendingMedia, mediaData{
				PostID:    postID,
				MediaKey:  mediaKey,
				MediaType: "image",
				Metadata:  metadataJSON,
			})
		}
	}

	// Batch Insert Media
	for i := 0; i < len(pendingMedia); i += postsBatchSize {
		end := i + postsBatchSize
		if end > len(pendingMedia) {
			end = len(pendingMedia)
		}
		insertPostMediaBatch(tx, pendingMedia[i:end])
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit posts failed: %v", err)
	}

	log.Printf("Seeded: %d Posts (with media)", len(allPostIDs))
	return allPostIDs
}

func insertPostsBatch(tx *sqlx.Tx, posts []postData) []int64 {
	if len(posts) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(posts))
	valueArgs := make([]interface{}, 0, len(posts)*3)

	for i, p := range posts {
		// $1, $2, $3 ... $4, $5, $6
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, p.UserID, p.Content, p.Visibility)
	}

	stmt := fmt.Sprintf("INSERT INTO posts (user_id, content, visibility) VALUES %s RETURNING id", strings.Join(valueStrings, ","))

	rows, err := tx.Query(stmt, valueArgs...)
	if err != nil {
		log.Panicf("batch insert posts failed: %v", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Panicf("scan post id failed: %v", err)
		}
		ids = append(ids, id)
	}
	return ids
}

func insertPostMediaBatch(tx *sqlx.Tx, media []mediaData) {
	if len(media) == 0 {
		return
	}

	valueStrings := make([]string, 0, len(media))
	valueArgs := make([]interface{}, 0, len(media)*4)

	for i, m := range media {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		valueArgs = append(valueArgs, m.PostID, m.MediaKey, m.MediaType, m.Metadata)
	}

	stmt := fmt.Sprintf("INSERT INTO post_media (post_id, media_key, media_type, metadata) VALUES %s", strings.Join(valueStrings, ","))

	if _, err := tx.Exec(stmt, valueArgs...); err != nil {
		log.Panicf("batch insert post_media failed: %v", err)
	}
}

func TruncatePosts(db *sqlx.DB) {
	_, err := db.Exec(`TRUNCATE TABLE posts RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Panicf("cannot clean posts data: %v", err)
	}
}
