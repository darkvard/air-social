package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/domain/user"
	"air-social/internal/repository/postgres"
)

type RepoProvider struct {
	User user.Repository
}

func NewRepoProvider(db *sqlx.DB) *RepoProvider {
	return &RepoProvider{
		User: postgres.NewUserRepoImpl(db),
	}
}
