package main

import (
	"log"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"

	"air-social/cmd/seed/config"
	"air-social/cmd/seed/db"
	"air-social/cmd/seed/modules"
)

func main() {
	start := time.Now()
	log.Println("🌱 Starting database seeding...")

	// 1. Initial setup
	gofakeit.Seed(0)
	cfg := config.Load()
	conn := db.Connect()
	defer conn.Close()

	// 2. Clean database
	truncateData(conn)

	// 3. Seed new data
	seedData(conn, cfg)

	log.Printf("✅ Seeding finished in %s", time.Since(start))
}

// truncateData cleans all relevant tables in the correct order to respect foreign key constraints.
func truncateData(db *sqlx.DB) {
	log.Println("🗑️  Cleaning existing data...")
	// The order is important. We start from tables that depend on others.
	modules.TruncateComments(db)
	modules.TruncatePosts(db)
	// Truncating 'users' will cascade to 'follows', 'post_likes', 'comment_likes' etc.
	// due to `ON DELETE CASCADE` in the database schema.
	modules.TruncateUser(db)
	log.Println("✅ Data cleaned.")
}

// seedData populates the database with new records.
// It runs independent tasks in parallel while maintaining the correct sequence for dependent tasks.
func seedData(db *sqlx.DB, cfg config.SeedConfig) {
	log.Println("🌱 Seeding new data...")

	// --- Sequential Step 1: Users must exist first ---
	userIDs := modules.SeedUsers(db, cfg.Users.Total)

	var wg sync.WaitGroup

	// --- Parallel Branch 1: Social Graph (Follows) ---
	// This branch only depends on users, so it can run in parallel with the content branch.
	wg.Go(func() {
		modules.SeedFollows(db, userIDs, cfg.Follows.PerUser)
	})

	// --- Sequential Branch 2: Content Graph ---
	// This branch has its own internal sequential dependencies.
	postIDs := modules.SeedPosts(db, userIDs, cfg)
	commentIDs := modules.SeedComments(db, postIDs, userIDs, cfg)
	// Likes depend on posts and comments
	modules.SeedLikes(db, postIDs, commentIDs, userIDs, cfg)

	// Wait for all parallel branches to complete
	wg.Wait()

	// --- Final Step: Calculate and seed aggregate stats ---
	// This must run last, after all source data (likes, comments, shares) is created.
	modules.SeedStats(db)
}
