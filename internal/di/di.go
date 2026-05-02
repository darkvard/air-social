package di

import (
	"net/http"

	"air-social/internal/config"
	"air-social/internal/di/provider"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/server"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/ws"
)

type Container struct {
	Server *http.Server
	Worker *worker.Manager
	Hub    *ws.Hub
	Infra  *provider.Infrastructure
}

func Initialize(cfg config.Config) (*Container, func(), error) {
	infra, cleanup, err := provider.NewInfrastructure(cfg)
	if err != nil {
		return nil, nil, err
	}

	handleError := func(err error) (*Container, func(), error) {
		cleanup()
		return nil, nil, err
	}

	adapter, err := provider.NewAdapter(cfg, infra)
	if err != nil {
		return handleError(err)
	}

	prov := provider.NewProvider(cfg, adapter)
	repo := provider.NewRepository(infra)
	usecase := provider.NewUseCase(provider.UseCaseDeps{
		Cfg:     cfg,
		Infra:   infra,
		Prov:    prov,
		Repo:    repo,
		Adapter: adapter,
	})
	handler := provider.NewHandler(prov, usecase)
	middleware := middleware.NewManager(cfg.Server, cfg.Limiter, prov.Token, adapter.RateLimiter)
	hub := ws.NewHub(infra.Redis, adapter.Presence)
	server := server.NewServer(cfg, prov, handler, middleware, hub)

	workers := provider.NewWorkers(infra, adapter, prov, usecase)

	return &Container{
		Server: server,
		Worker: workers,
		Hub:    hub,
		Infra:  infra,
	}, cleanup, nil
}
