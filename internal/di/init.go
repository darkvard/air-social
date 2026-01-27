package di

import (
	"net/http"

	"air-social/internal/config"
	transport "air-social/internal/transport/http"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/ws"
)

type Container struct {
	Server *http.Server
	Worker *worker.Manager
	Hub    *ws.Hub
	Infra  *Infrastructures
}

func Initialize(cfg config.Config) (*Container, func(), error) {
	url := transport.NewURLFactory(cfg.Server)
	url.PrintInfraConsole()

	infrastructures, cleanup, err := initInfrastructures(cfg)
	if err != nil {
		return nil, nil, err
	}

	handleError := func(err error) (*Container, func(), error) {
		cleanup()
		return nil, nil, err
	}

	adapters, err := initAdapters(cfg, infrastructures)
	if err != nil {
		return handleError(err)
	}

	repositories := initRepository(infrastructures)
	services := initServices(cfg, url, infrastructures, repositories, adapters)
	handlers := initHandlers(services)
	middlewares := middleware.NewManager(cfg.Server, services.Token)

	server := transport.NewServer(cfg, url, middlewares, handlers.Auth, handlers.User, handlers.Media, handlers.Health)

	return &Container{
		Server: server,
		Worker: initWorkers(infrastructures, adapters, services),
		Hub:    ws.NewHub(),
		Infra:  infrastructures,
	}, cleanup, nil
}
