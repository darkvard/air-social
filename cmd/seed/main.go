package main

import (
	"air-social/cmd/seed/config"
	"air-social/cmd/seed/db"
	"air-social/cmd/seed/modules"
)

func main() {
	cfg := config.Load()
	conn := db.Connect()
	defer conn.Close()

	modules.TruncateUser(conn)
	users := modules.SeedUsers(conn, cfg.Users.Total)
	modules.SeedFollows(conn, users, cfg.Follows.PerUser)
	modules.SeedPosts(conn, users, cfg.Posts.PerUser)
}
