package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/internal/config"
	"air-social/pkg"
)

type HttpProvider struct {
	Repo    *RepoProvider
	Service *ServiceProvider
	Handler *HandlerProvider
}

func NewHttpProvider(db *sqlx.DB, cfg config.TokenConfig, hash pkg.Hasher) *HttpProvider {
	repo := NewRepoProvider(db)
	service := NewServiceProvider(repo, cfg, hash)
	handler := NewHandlerProvider(service)
	return &HttpProvider{
		Repo:    repo,
		Service: service,
		Handler: handler,
	}
}
