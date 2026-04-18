package main

import (
	"log"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/v2/mongo"

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
	mongoDB := db.ConnectMongo()

	// 2. Clean database
	truncateData(conn, mongoDB)

	// 3. Seed new data
	seedData(conn, mongoDB, cfg)

	log.Printf("✅ Seeding finished in %s", time.Since(start))
}

// truncateData cleans all relevant tables in the correct order to respect foreign key constraints.
func truncateData(db *sqlx.DB, mongoDB *mongo.Database) {
	log.Println("🗑️  Cleaning existing data...")
	modules.TruncateConversations(mongoDB)
	modules.TruncateComments(db)
	modules.TruncatePosts(db)
	modules.TruncateUser(db)
	log.Println("✅ Data cleaned.")
}

// seedData populates the database with new records.
// It runs independent tasks in parallel while maintaining the correct sequence for dependent tasks.
func seedData(db *sqlx.DB, mongoDB *mongo.Database, cfg config.SeedConfig) {
	log.Println("🌱 Seeding new data...")

	// --- Sequential Step 1: Users must exist first ---
	userIDs := modules.SeedUsers(db, cfg.Users.Total)

	var wg sync.WaitGroup

	// --- Parallel Branch 1: Social Graph (Follows) + Conversations ---
	// Both depend only on users and can run in parallel with the content branch.
	wg.Go(func() {
		modules.SeedFollows(db, userIDs, cfg.Follows.PerUser)
	})
	wg.Go(func() {
		modules.SeedConversations(mongoDB, userIDs, cfg)
	})

	// --- Sequential Branch 2: Content Graph ---
	// This branch has its own internal sequential dependencies.
	postIDs := modules.SeedPosts(db, userIDs, cfg)
	commentIDs := modules.SeedComments(db, postIDs, userIDs, cfg)
	modules.SeedLikes(db, postIDs, commentIDs, userIDs, cfg)

	// Wait for all parallel branches to complete
	wg.Wait()

	// --- Final Step: Calculate and seed aggregate stats ---
	modules.SeedStats(db)
}
