package di

import (
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/mailer"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/infrastructure/redis"
	"air-social/internal/service"
)

type ServiceContainer struct {
	Health service.HealthService
	Auth   service.AuthService
	Token  service.TokenService
	User   service.UserService
	Media  service.MediaService
	Email  service.EmailService

	Event domain.EventPublisher
	Cache domain.CacheStorage
	File  domain.FileStorage
}

func NewServices(cfg config.Config, ifc *InfraContainer, url domain.URLFactory) (*ServiceContainer, error) {
	// infra
	fileStorage, err := minio.NewMinioStorage(ifc.Minio)
	if err != nil {
		return nil, err
	}
	cacheCache, err := redis.NewRedisCache(ifc.Redis)
	if err != nil {
		return nil, err
	}
	eventPublisher, err := rabbitmq.NewEventPublisher(ifc.Rabbit)
	if err != nil {
		return nil, err
	}

	// repos
	userRepository := postgres.NewUserRepository(ifc.DB)
	tokenRepository := postgres.NewTokenRepository(ifc.DB)

	// services
	mediaService := service.NewMediaService(fileStorage, cacheCache, domain.FileConfig{
		PublicPathPrefix: url.FileStorageBaseURL(),
		BucketPublic:     cfg.MinIO.BucketPublic,
		BucketPrivate:    cfg.MinIO.BucketPrivate,
	})
	healthService := service.NewHealthService(ifc.DB, ifc.Redis, &rabbitmq.HealthChecker{
		Conn: ifc.Rabbit,
		URL:  cfg.RabbitMQ.URL,
	}, ifc.Minio, url)
	tokenService := service.NewTokenService(tokenRepository, cfg.Token)
	userService := service.NewUserService(userRepository, mediaService)
	authService := service.NewAuthService(userService, tokenService, url, eventPublisher, cacheCache)
	emailService := service.NewEmailService(mailer.NewMailtrap(cfg.Mailer))

	return &ServiceContainer{
		Health: healthService,
		Auth:   authService,
		Token:  tokenService,
		User:   userService,
		Media:  mediaService,
		Email:  emailService,

		Event: eventPublisher,
		Cache: cacheCache,
		File:  fileStorage,
	}, nil
}
