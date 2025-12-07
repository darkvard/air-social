package provider

import (
	"github.com/jmoiron/sqlx"
	rc "github.com/redis/go-redis/v9"

	"air-social/internal/cache"
	"air-social/internal/domain"
	"air-social/internal/repository/postgres"
	"air-social/internal/repository/redis"
)

type RepoProvider struct {
	User  domain.UserRepository
	Token domain.TokenRepository
	Cache cache.CacheStorage
}

func NewRepoProvider(db *sqlx.DB, rc *rc.Client) *RepoProvider {
	return &RepoProvider{
		User:  postgres.NewUserRepoImpl(db),
		Token: postgres.NewTokenRepository(db),
		Cache: redis.NewRedisCache(rc),
	}
}
