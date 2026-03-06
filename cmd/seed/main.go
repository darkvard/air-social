package main

import (
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"air-social/cmd/seed/config"
	"air-social/cmd/seed/db"
	"air-social/cmd/seed/modules"
)

func main() {
	start := time.Now()
	log.Println("🌱 Starting database seeding...")

	// 1. Setup
	gofakeit.Seed(0)
	cfg := config.Load()
	conn := db.Connect()
	defer conn.Close()

	// 2. Truncate data in reverse order of dependency
	log.Println("🗑️  Cleaning existing data...")
	modules.TruncateComments(conn)
	modules.TruncatePosts(conn)
	modules.TruncateUser(conn) // This will cascade to follows, likes, etc.
	log.Println("✅ Data cleaned.")

	// 3. Seed data in order of dependency
	userIDs := modules.SeedUsers(conn, cfg.Users.Total)
	modules.SeedFollows(conn, userIDs, cfg.Follows.PerUser)
	postIDs := modules.SeedPosts(conn, userIDs, cfg)
	commentIDs := modules.SeedComments(conn, postIDs, userIDs, cfg)
	modules.SeedLikes(conn, postIDs, commentIDs, userIDs, cfg)

	log.Printf("✅ Seeding finished in %s", time.Since(start))
}
