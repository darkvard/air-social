package di

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infra/msg"
	"air-social/internal/infra/storage"
	"air-social/internal/repository/postgres"
	"air-social/internal/repository/redis"
	"air-social/internal/service"
)

type ServiceContainer struct {
	Event domain.EventPublisher
	Cache domain.CacheStorage
	File  domain.FileStorage

	Auth  service.AuthService
	User  service.UserService
	Media service.MediaService
	Token service.TokenService
}

func NewServices(cfg config.Config, ifc *InfraContainer, url domain.URLFactory) (*ServiceContainer, error) {
	// file
	fileStorage := storage.NewMinioStorage(ifc.Minio)
	fileConfig := domain.FileConfig{
		PublicURL:     cfg.MinIO.PublicURL,
		BucketPublic:  cfg.MinIO.BucketPublic,
		BucketPrivate: cfg.MinIO.BucketPrivate,
	}

	// cache
	if ifc.Redis == nil {
		return nil, errors.New("redis cannot nil")
	}
	cacheStorage := redis.NewRedisCache(ifc.Redis)

	// event
	if ifc.Rabbit == nil {
		return nil, errors.New("rabbit-mq cannot nil")
	}
	eventQueue := newPublisher(ifc.Rabbit)

	// repos
	userRepo := postgres.NewUserRepoImpl(ifc.DB)
	tokenRepo := postgres.NewTokenRepository(ifc.DB)

	// services
	tokenSvc := service.NewTokenService(tokenRepo, cfg.Token)
	mediaSvc := service.NewMediaService(fileStorage, cacheStorage, fileConfig)
	userSvc := service.NewUserService(userRepo, mediaSvc)
	authSvc := service.NewAuthService(userSvc, tokenSvc, url, eventQueue, cacheStorage)

	return &ServiceContainer{
		Event: eventQueue,
		Cache: cacheStorage,
		File:  fileStorage,

		Auth:  authSvc,
		User:  userSvc,
		Media: mediaSvc,
		Token: tokenSvc,
	}, nil
}

func newPublisher(conn *amqp.Connection) *msg.Publisher {
	pub, err := msg.NewPublisher(
		conn,
		msg.EventsExchange,
		10,
	)
	if err != nil {
		panic(err)
	}
	return pub
}
