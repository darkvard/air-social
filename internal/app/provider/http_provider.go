package provider

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/pkg"
)

type HttpProvider struct {
	Repo    *RepoProvider
	Service *ServiceProvider
	Handler *HandlerProvider
}

func NewHttpProvider(
	db *sqlx.DB,
	cfg config.TokenConfig,
	hash pkg.Hasher,
	cache *redis.Client,
	queue domain.EventPublisher,
) *HttpProvider {
	repo := NewRepoProvider(db, cache)
	service := NewServiceProvider(repo, cfg, hash, queue)
	handler := NewHandlerProvider(service)
	return &HttpProvider{
		Repo:    repo,
		Service: service,
		Handler: handler,
	}
}
