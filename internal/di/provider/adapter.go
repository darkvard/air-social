package provider

import (
	"air-social/internal/config"
	"air-social/internal/domain/common"
	"air-social/internal/domain/media"
	"air-social/internal/infrastructure/mailer"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/infrastructure/redis"
)

type Adapter struct {
	Media    media.Storage
	Cache    common.Cache
	EventPub common.EventPublisher
	Mailer   common.Mailer
}

func NewAdapter(cfg config.Config, infra *Infrastructure) (Adapter, error) {
	var empty Adapter

	fileStorage, err := minio.NewStorage(infra.Minio)
	if err != nil {
		return empty, err
	}

	cache, err := redis.NewCache(infra.Redis)
	if err != nil {
		return empty, err
	}

	eventPub, err := rabbitmq.NewEventPublisher(infra.Rabbit)
	if err != nil {
		return empty, err
	}

	mailer := mailer.NewMailtrap(cfg.Mailer)

	return Adapter{
		Media:    fileStorage,
		Cache:    cache,
		EventPub: eventPub,
		Mailer:   mailer,
	}, nil
}
