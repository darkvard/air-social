package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/routes"
)

type HttpProvider struct {
	Repository *RepoProvider
	Service    *ServiceProvider
	Handler    *HandlerProvider
}

func NewHttpProvider(
	db *sqlx.DB,
	cfg config.TokenConfig,
	cs cache.CacheStorage,
	pub domain.EventPublisher,
	rr routes.Registry,
) *HttpProvider {
	repository := NewRepoProvider(db, cs)
	service := NewServiceProvider(repository, cfg, pub, rr, cs)
	handler := NewHandlerProvider(service)
	return &HttpProvider{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
