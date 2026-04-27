package provider

import (
	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/domain/limit"
	"air-social/internal/domain/media"
	"air-social/internal/infrastructure/mailer"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/rabbitmq"
	rediscache "air-social/internal/infrastructure/redis/cache"
	redislimit "air-social/internal/infrastructure/redis/limit"
	redismessage "air-social/internal/infrastructure/redis/message"
)

type Adapter struct {
	Media          media.Storage
	AuthCache      cache.AtomicCache[string]
	HashStore      cache.HashStore
	SortedSetStore cache.SortedSetStore
	EventPub       common.EventPublisher
	Mailer         common.Mailer
	RateLimiter    limit.RateLimiter
	Presence       chat.PresenceStore
	Unread         chat.UnreadStore
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
		AuthCache:      rediscache.NewRedisStore[string](infra.Redis),
		HashStore:      rediscache.NewRedisHashStore(infra.Redis),
		SortedSetStore: rediscache.NewRedisSortedSetStore(infra.Redis),
		EventPub:       eventPub,
		Mailer:         mailer,
		RateLimiter:    redislimit.NewRedisRateLimiter(infra.Redis),
		Presence:       redismessage.NewPresenceStorage(infra.Redis),
		Unread:         redismessage.NewUnreadStorage(infra.Redis),
	}, nil
}
