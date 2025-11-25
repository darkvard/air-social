package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"air-social/internal/config"
)

func NewDatabase(ps config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", ps.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open DB: %w", err)
	}

	db.SetMaxOpenConns(ps.MaxOpenConn)
	db.SetMaxIdleConns(ps.MaxIdleConn)
	db.SetConnMaxLifetime(ps.MaxLifeTime)
	db.SetConnMaxIdleTime(ps.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("cannot ping DB: %w", err)
	}

	return db, nil
}
