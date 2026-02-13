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
	Infra  *provider.Infrastructures
}

func Initialize(cfg config.Config) (*Container, func(), error) {
	url := route.NewURLFactory(cfg)
	url.PrintInfraConsole()

	infrastructures, cleanup, err := provider.NewInfrastructures(cfg)
	if err != nil {
		return nil, nil, err
	}

	handleError := func(err error) (*Container, func(), error) {
		cleanup()
		return nil, nil, err
	}

	adapters, err := provider.NewAdapters(cfg, infrastructures)
	if err != nil {
		return handleError(err)
	}

	repositories := provider.NewRepositories(infrastructures)
	services := provider.NewServices(cfg, url, infrastructures, repositories, adapters)
	handlers := provider.NewHandlers(services, url)
	middlewares := middleware.NewManager(cfg.Server, services.Token)
	workers := provider.NewWorkers(infrastructures, adapters, services)
	server := server.NewServer(cfg, url, middlewares, handlers)

	return &Container{
		Server: server,
		Worker: workers,
		Hub:    ws.NewHub(),
		Infra:  infrastructures,
	}, cleanup, nil
}
