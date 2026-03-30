package provider

import (
	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain/common"
	limitdomain "air-social/internal/domain/limit"
	"air-social/internal/domain/media"
	"air-social/internal/infrastructure/mailer"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/infrastructure/redis"
)

type Adapter struct {
	Media          media.Storage
	AuthCache      cache.AtomicCache[string]
	HashStore      cache.HashStore
	SortedSetStore cache.SortedSetStore
	EventPub       common.EventPublisher
	Mailer         common.Mailer
	RateLimiter    limitdomain.RateLimiter
}

func NewAdapter(cfg config.Config, infra *Infrastructure) (Adapter, error) {
	var empty Adapter

	fileStorage, err := minio.NewStorage(infra.Minio)
	if err != nil {
		return empty, err
	}

	eventPub, err := rabbitmq.NewEventPublisher(infra.Rabbit)
	if err != nil {
		return empty, err
	}

	mailer := mailer.NewMailtrap(cfg.Mailer)

	return Adapter{
		Media:          fileStorage,
		AuthCache:      redis.NewRedisStore[string](infra.Redis),
		HashStore:      redis.NewRedisHashStore(infra.Redis),
		SortedSetStore: redis.NewRedisSortedSetStore(infra.Redis),
		EventPub:       eventPub,
		Mailer:         mailer,
		RateLimiter:    redis.NewRedisRateLimiter(infra.Redis),
	}, nil
}
