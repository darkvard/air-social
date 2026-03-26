package provider

import (
	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
	"air-social/internal/domain/auth/verify"
	"air-social/internal/domain/common"
	"air-social/internal/domain/feed"
	"air-social/internal/domain/stats"
	"air-social/internal/infrastructure/url"
)

type Provider struct {
	Link   common.AppLinkManager
	Token  token.Provider
	Verify verify.Provider
	Stats  stats.Cache
	Feed   feed.Cache
}

func NewProvider(cfg config.Config, adapter Adapter) Provider {
	linkManager := newLinkManager(cfg)
	tokenProvider := token.NewProvider(token.Deps{Config: cfg.Token, Cache: adapter.AuthCache})
	verifyProvider := newVerifyProvider(adapter, linkManager.LinkProvider)
	statsProvider := newStatsProvider(adapter)
	feedCacheProvider := newFeedCacheProvider(adapter)

	return Provider{
		Link:   linkManager,
		Token:  tokenProvider,
		Verify: verifyProvider,
		Stats:  statsProvider,
		Feed:   feedCacheProvider,
	}
}

func newLinkManager(cfg config.Config) common.AppLinkManager {
	manager := url.NewManager(cfg)
	return common.AppLinkManager{
		SystemProvider: manager.System,
		RouteProvider:  manager.Route,
		LinkProvider:   manager.Link,
	}
}

func newVerifyProvider(adapter Adapter, link common.LinkProvider) verify.Provider {
	return verify.NewVerifyProvider(verify.Deps{
		Cache: adapter.AuthCache,
		Event: adapter.EventPub,
		Link:  link,
	})
}

func newStatsProvider(adapter Adapter) stats.Cache {
	return stats.NewCache(adapter.HashStore)
}

func newFeedCacheProvider(adapter Adapter) feed.Cache {
	return feed.NewCache(adapter.SortedSetStore)
}
