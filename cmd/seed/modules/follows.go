package modules

import (
	"log"

	"github.com/jmoiron/sqlx"
)

func SeedFollows(db *sqlx.DB, users []int64, perUser int) {
	// Safety: Rollback on panic/error; ignored if Commit succeeds
	tx := db.MustBegin()
	defer tx.Rollback()

	n := len(users)
	if n < 2 {
		return
	}

	// Optimization: Clamp perUser to max possible (n-1) to avoid redundant loops
	if perUser >= n {
		perUser = n - 1
	}

	for i, followeeID := range users {
		for j := 1; j <= perUser; j++ {
			// Algorithm: Circular wrap-around ensuring valid index [0, n-1]
			followerIndex := (i + j) % n
			followerID := users[followerIndex]

			_, err := tx.Exec(`
				INSERT INTO follows (follower_id, followee_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, followerID, followeeID)

			if err != nil {
				log.Panicf("seed follows query failed: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("seed follows commit failed: %v", err)
	}
}
