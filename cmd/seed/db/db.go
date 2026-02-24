package db

import (
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func Connect() *sqlx.DB {
	// Use Overload to prioritize .env.local over environment variables set by Make or Shell
	_ = godotenv.Overload(".env.local")

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSL_MODE"),
		)
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
