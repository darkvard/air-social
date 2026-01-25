package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"air-social/internal/config"
	"air-social/pkg"
)

func NewConnection(ps config.PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", ps.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot open DB: %w", err)
	}

	db.SetMaxOpenConns(ps.MaxOpenConn)
	db.SetMaxIdleConns(ps.MaxIdleConn)
	db.SetConnMaxLifetime(ps.MaxLifeTime)
	db.SetConnMaxIdleTime(ps.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = pkg.Retry(ctx, 5, 2*time.Second, func() error {
		return db.PingContext(ctx)
	}); err != nil {
		return nil, fmt.Errorf("postgres: %w", err)

	}

	return db, nil
}
