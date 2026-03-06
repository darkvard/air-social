package modules

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"

	"air-social/cmd/seed/config"
)

// batchSize defines how many records to insert in a single query.
const likesBatchSize = 1000

func SeedLikes(db *sqlx.DB, postIDs, commentIDs, userIDs []int64, cfg config.SeedConfig) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Seed Post Likes using Batch Insert
	go func() {
		defer wg.Done()
		seedLikesFor(db, "post_likes", "post_id", postIDs, userIDs, cfg.Likes.PerPost)
	}()

	// Goroutine 2: Seed Comment Likes using Batch Insert
	go func() {
		defer wg.Done()
		seedLikesFor(db, "comment_likes", "comment_id", commentIDs, userIDs, cfg.Likes.PerComment)
	}()

	wg.Wait()
}

// seedLikesFor is a generic helper to perform batch inserts for likes.
func seedLikesFor(db *sqlx.DB, tableName, entityColumn string, entityIDs, userIDs []int64, likesPerEntity int) {
	// 1. Collect all like pairs first to avoid inserting one-by-one
	var likePairs [][2]int64
	uniqueLikes := make(map[string]struct{})

	// Create a mutable copy of userIDs for shuffling
	shuffledUserIDs := make([]int64, len(userIDs))
	copy(shuffledUserIDs, userIDs)

	for _, entityID := range entityIDs {
		rand.Shuffle(len(shuffledUserIDs), func(i, j int) { shuffledUserIDs[i], shuffledUserIDs[j] = shuffledUserIDs[j], shuffledUserIDs[i] })
		numLikes := rand.Intn(likesPerEntity + 1)

		for i := 0; i < numLikes && i < len(shuffledUserIDs); i++ {
			userID := shuffledUserIDs[i]
			// Use a map to ensure a user doesn't like the same entity twice in the seed data
			key := fmt.Sprintf("%d-%d", entityID, userID)
			if _, exists := uniqueLikes[key]; !exists {
				likePairs = append(likePairs, [2]int64{entityID, userID})
				uniqueLikes[key] = struct{}{}
			}
		}
	}

	if len(likePairs) == 0 {
		log.Printf("Seeded: 0 %s", tableName)
		return
	}

	tx := db.MustBegin()
	defer tx.Rollback()

	// 2. Insert the collected pairs in batches
	totalInserted := 0
	for i := 0; i < len(likePairs); i += likesBatchSize {
		end := i + likesBatchSize
		if end > len(likePairs) {
			end = len(likePairs)
		}
		batch := likePairs[i:end]

		if len(batch) == 0 {
			continue
		}

		// Build the multi-value INSERT statement for the current batch
		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*2)
		argCounter := 1
		for _, pair := range batch {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", argCounter, argCounter+1))
			valueArgs = append(valueArgs, pair[0], pair[1])
			argCounter += 2
		}

		stmt := fmt.Sprintf("INSERT INTO %s (%s, user_id) VALUES %s ON CONFLICT (%s, user_id) DO NOTHING", tableName, entityColumn, strings.Join(valueStrings, ","), entityColumn)
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			log.Panicf("batch insert for %s failed: %v", tableName, err)
		}
		totalInserted += len(batch)
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit for %s failed: %v", tableName, err)
	}
	log.Printf("Seeded: %d %s", totalInserted, tableName)
}
