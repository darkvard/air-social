package url

import (
	"air-social/internal/config"
	"air-social/internal/domain/common"
)

type Manager struct {
	System common.SystemProvider
	Route  common.RouteProvider
	Link   common.LinkProvider
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
