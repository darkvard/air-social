package url

import (
	"fmt"

	"air-social/internal/config"
)

type route struct {
	protocol string
	domain   string
	appName  string
	version  string
}

func newRouteProvider(cfg config.ServerConfig) *route {
	return &route{
		protocol: cfg.Protocol,
		domain:   cfg.Domain,
		appName:  cfg.AppName,
		version:  cfg.Version,
	}
}

func (u *route) BaseURL() string {
	return fmt.Sprintf("%s://%s", u.protocol, u.domain)
}

func (u *route) ApiPath() string {
	return fmt.Sprintf("%s/%s/api/%s", u.BaseURL(), u.appName, u.version)
}
