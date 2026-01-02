package provider

import (
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	goredis "github.com/redis/go-redis/v9"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/routes"
)

type Container struct {
	DB       *sqlx.DB
	Redis    *goredis.Client
	Rabbit   *amqp.Connection
	Config   *config.Config
	Cache    cache.CacheStorage
	Pub      domain.EventPublisher
	Registry routes.Registry
}

type HttpProvider struct {
	Repository *RepoProvider
	Service    *ServiceProvider
	Handler    *HandlerProvider
}

func NewHttpProvider(deps *Container) *HttpProvider {
	repository := NewRepoProvider(deps.DB, deps.Cache)
	service := NewServiceProvider(repository, deps.Config.Token, deps.Pub, deps.Registry, deps.Cache)
	handler := NewHandlerProvider(deps, service)
	return &HttpProvider{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
