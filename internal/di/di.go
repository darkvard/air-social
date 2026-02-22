package di

import (
	"net/http"

	"air-social/internal/config"
	"air-social/internal/di/provider"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/route"
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
	url := route.NewURLFactory(cfg)
	url.PrintInfraConsole()

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
	handler := provider.NewHandlers(prov, usecase)

	workers := provider.NewWorkers(infra, adapter)
	middleware := middleware.NewManager(cfg.Server, prov.Token)

	server := server.NewServer(cfg, url, middleware, handler)

	return &Container{
		Server: server,
		Worker: workers,
		Hub:    ws.NewHub(),
		Infra:  infra,
	}, cleanup, nil
}
