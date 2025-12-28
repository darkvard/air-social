package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/cache"
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
	cache cache.CacheStorage,
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
