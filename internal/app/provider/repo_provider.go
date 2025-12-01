package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/domain"
	"air-social/internal/repository/postgres"
)

type RepoProvider struct {
	User  domain.UserRepository
	Token domain.TokenRepository
}

func NewRepoProvider(db *sqlx.DB) *RepoProvider {
	return &RepoProvider{
		User:  postgres.NewUserRepoImpl(db),
		Token: postgres.NewTokenRepository(db),
	}
}
