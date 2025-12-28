package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/cache"
	"air-social/internal/domain"
	"air-social/internal/repository/postgres"
)

type RepoProvider struct {
	User  domain.UserRepository
	Token domain.TokenRepository
	Cache cache.CacheStorage
}

func NewRepoProvider(db *sqlx.DB, cs cache.CacheStorage) *RepoProvider {
	return &RepoProvider{
		User:  postgres.NewUserRepoImpl(db),
		Token: postgres.NewTokenRepository(db),
		Cache: cs,
	}
}
