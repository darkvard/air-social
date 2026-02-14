package modules

import (
	"fmt"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
)

func SeedPosts(db *sqlx.DB, userIDs []int64, postsPerUser int) {
	tx := db.MustBegin()
	defer tx.Rollback()

	postStmt := preparePostStmt(tx)
	mediaStmt := prepareMediaStmt(tx)

	for _, userID := range userIDs {
		for range postsPerUser {
			postID := execInsertPost(postStmt, userID)
			seedMediaForPost(mediaStmt, postID)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("seed posts commit failed: %v", err)
	}
	
	log.Printf("Seeded: %d Posts (%d per user)", len(userIDs)*postsPerUser, postsPerUser)
}

// --- Internal Helpers ---

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

func prepareMediaStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
		INSERT INTO post_media (post_id, media_key, media_type, metadata)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		log.Panicf("prepare media stmt failed: %v", err)
	}
	return stmt
}

func execInsertPost(stmt *sqlx.Stmt, userID int64) int64 {
	var postID int64
	visibility := getRandVisibility()
	content := gofakeit.Sentence(15)

	if err := stmt.QueryRow(userID, content, visibility).Scan(&postID); err != nil {
		log.Panicf("insert post failed: %v", err)
	}
	return postID
}

func seedMediaForPost(stmt *sqlx.Stmt, postID int64) {
	if gofakeit.Number(1, 10) > 7 {
		numMedia := gofakeit.Number(1, 4)
		for range numMedia {
			mType, ext := getRandMediaType()
			meta := generateMeta(ext)
			
			if _, err := stmt.Exec(postID, gofakeit.UUID(), mType, meta); err != nil {
				log.Panicf("insert media failed: %v", err)
			}
		}
	}
}

// --- Utilities ---

func getRandVisibility() string {
	if gofakeit.Number(1, 10) > 8 {
		return "followers"
	}
	return "public"
}

func getRandMediaType() (string, string) {
	if gofakeit.Number(1, 10) > 9 {
		return "video/mp4", "mp4"
	}
	return "image/jpeg", "jpg"
}

func generateMeta(ext string) string {
	return fmt.Sprintf(
		`{"width": %d, "height": %d, "size": %d, "file_name": "%s"}`,
		gofakeit.Number(800, 1920),
		gofakeit.Number(600, 1080),
		gofakeit.Number(100000, 5000000),
		gofakeit.Word()+"."+ext,
	)
}