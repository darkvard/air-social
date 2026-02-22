package url

import (
	"air-social/internal/config"
	"air-social/internal/domain/shared"
)

type Manager struct {
	System shared.SystemProvider
	Route  shared.RouteProvider
	Link   shared.LinkProvider
}

func NewManager(cfg config.Config) Manager {
	route := newRouteProvider(cfg.Server)
	system := newSystemProvider(route)
	link := newLinkProvider(route, cfg.MinIO)
	return Manager{
		System: system,
		Route:  route,
		Link:   link,
	}
}
