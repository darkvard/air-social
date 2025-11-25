package provider

import (
	"github.com/jmoiron/sqlx"

	"air-social/pkg"
)

type HttpProvider struct {
	Repo    *RepoProvider
	Service *ServiceProvider
	Handler *HandlerProvider
}

func NewHttpProvider(db *sqlx.DB, jwt pkg.JWTAuth, hash pkg.Hasher) *HttpProvider {
	repo := NewRepoProvider(db)
	service := NewServiceProvider(repo, jwt, hash)
	handler := NewHandlerProvider(service)
	return &HttpProvider{
		Repo:    repo,
		Service: service,
		Handler: handler,
	}
}
