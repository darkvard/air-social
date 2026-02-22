package provider

import (
	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
	"air-social/internal/domain/auth/verify"
	"air-social/internal/domain/shared"
	"air-social/internal/infrastructure/url"
)

type Provider struct {
	Link   shared.AppLinkManager
	Token  token.Provider
	Verify verify.Provider
}

func NewProvider(cfg config.Config, adapter Adapter) Provider {
	linkManager := newLinkManager(cfg)
	tokenProvider := token.NewProvider(cfg.Token, adapter.Cache)
	verifyProvider := newVerifyProvider(adapter, linkManager.LinkProvider)

	return Provider{
		Link:   linkManager,
		Token:  tokenProvider,
		Verify: verifyProvider,
	}
}

func newLinkManager(cfg config.Config) shared.AppLinkManager {
	manager := url.NewManager(cfg)
	return shared.AppLinkManager{
		SystemProvider: manager.System,
		RouteProvider:  manager.Route,
		LinkProvider:   manager.Link,
	}
}

func newVerifyProvider(adapter Adapter, link shared.LinkProvider) verify.Provider {
	return verify.NewVerifyProvider(verify.Deps{
		Cache: adapter.Cache,
		Event: adapter.EventPub,
		Link:  link,
	})
}
