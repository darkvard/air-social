package modules

import (
	"log"
	"math/rand"

	"github.com/jmoiron/sqlx"
)

func SeedFollows(db *sqlx.DB, userIDs []int64, perUser int) {
	tx := db.MustBegin()
	defer tx.Rollback()

	stmt := prepareFollowStmt(tx)

	totalFollows := 0
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

			_, err := stmt.Exec(followerID, followeeID)
			if err != nil {
				log.Panicf("insert follow failed: %v", err)
			}
			count++
			totalFollows++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Panicf("commit follows failed: %v", err)
	}

	log.Printf("Seeded: %d Follows", totalFollows)
}

func prepareFollowStmt(tx *sqlx.Tx) *sqlx.Stmt {
	stmt, err := tx.Preparex(`
        INSERT INTO follows (follower_id, followee_id)
        VALUES ($1, $2)
        ON CONFLICT (follower_id, followee_id) DO NOTHING
    `)
	if err != nil {
		log.Panicf("prepare follow stmt failed: %v", err)
	}
	return stmt
}
