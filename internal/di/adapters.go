package di

import (
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/mailer"
	minioInfra "air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/rabbitmq"
	redisInfra "air-social/internal/infrastructure/redis"
)

type Adapters struct {
	FileStorage domain.FileStorage
	Cache       domain.CacheStorage
	EventPub    domain.EventPublisher
	Mailer      domain.Mailer
}

func initAdapters(cfg config.Config, infra *Infrastructures) (*Adapters, error) {
	fileStorage, err := minioInfra.NewMinioStorage(infra.Minio)
	if err != nil {
		return nil, err
	}

	cache, err := redisInfra.NewRedisCache(infra.Redis)
	if err != nil {
		return nil, err
	}

	eventPub, err := rabbitmq.NewEventPublisher(infra.Rabbit)
	if err != nil {
		return nil, err
	}

	mailer := mailer.NewMailtrap(cfg.Mailer)

	return &Adapters{
		FileStorage: fileStorage,
		Cache:       cache,
		EventPub:    eventPub,
		Mailer:      mailer,
	}, nil
}
