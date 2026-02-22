package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Health struct {
	db *sqlx.DB
}

func NewHealth(db *sqlx.DB) *Health {
	return &Health{
		db: db,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	return h.db.Ping()
}
