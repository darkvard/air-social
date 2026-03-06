package modules

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/jmoiron/sqlx"
)

const followsBatchSize = 1000

func SeedFollows(db *sqlx.DB, userIDs []int64, perUser int) {
	// 1. Collect all follow pairs first to avoid inserting one-by-one
	var followPairs [][2]int64
	uniqueFollows := make(map[string]struct{})

	// Create a mutable copy of userIDs for shuffling
	targets := make([]int64, len(userIDs))
	copy(targets, userIDs)

	for _, followerID := range userIDs {
		rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })

		count := 0
		for _, followeeID := range targets {
			if count >= perUser {
				break
			}

			if followerID == followeeID {
				continue
			}

			// Use a map to ensure a user doesn't follow the same person twice in the seed data
			key := fmt.Sprintf("%d-%d", followerID, followeeID)
			if _, exists := uniqueFollows[key]; !exists {
				followPairs = append(followPairs, [2]int64{followerID, followeeID})
				uniqueFollows[key] = struct{}{}
			}
			count++
		}
	}

	if len(followPairs) == 0 {
		log.Println("Seeded: 0 Follows")
		return
	}

	tx := db.MustBegin()
	defer tx.Rollback()

	// 2. Insert the collected pairs in batches
	totalInserted := 0
	for i := 0; i < len(followPairs); i += followsBatchSize {
		end := i + followsBatchSize
		if end > len(followPairs) {
			end = len(followPairs)
		}
		batch := followPairs[i:end]

		if len(batch) == 0 {
			continue
		}

		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*2)
		for j, pair := range batch {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", j*2+1, j*2+2))
			valueArgs = append(valueArgs, pair[0], pair[1])
		}
		stmt := fmt.Sprintf("INSERT INTO follows (follower_id, followee_id) VALUES %s ON CONFLICT (follower_id, followee_id) DO NOTHING", strings.Join(valueStrings, ","))
		_, err := tx.Exec(stmt, valueArgs...)
		if err != nil {
			log.Panicf("batch insert for follows failed: %v", err)
		}
		totalInserted += len(batch)
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit follows failed: %v", err)
	}

	log.Printf("Seeded: %d Follows", totalInserted)
}
