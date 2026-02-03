package di

import (
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/service"
)

type Services struct {
	Media  service.MediaService
	Health service.HealthService
	Token  service.TokenService
	User   service.UserService
	Auth   service.AuthService
	Email  service.EmailService
	Verify service.VerifyService
	Follow service.FollowService
}

func initServices(
	cfg config.Config,
	url domain.URLFactory,
	infra *Infrastructures,
	repository *Repositories,
	adapter *Adapters,
) *Services {

	mediaSvc := service.NewMediaService(adapter.FileStorage, domain.FileConfig{
		DomainPublic:  url.FileStorageBaseURL(),
		BucketPublic:  cfg.MinIO.BucketPublic,
		BucketPrivate: cfg.MinIO.BucketPrivate,
	})

	healthSvc := service.NewHealthService(infra.DB, infra.Redis, &rabbitmq.HealthChecker{
		Conn: infra.Rabbit,
		URL:  cfg.RabbitMQ.URL,
	}, infra.Minio, url)

	tokenSvc := service.NewTokenService(repository.Token, cfg.Token)
	verifySvc := service.NewVerifyService(adapter.Cache, adapter.EventPub, url)
	userSvc := service.NewUserService(repository.User, mediaSvc)
	authSvc := service.NewAuthService(userSvc, tokenSvc, verifySvc, adapter.Cache)
	emailSvc := service.NewEmailService(adapter.Mailer)
	followSvc := service.NewFollowService(repository.Follow)

	return &Services{
		Media:  mediaSvc,
		Health: healthSvc,
		Token:  tokenSvc,
		User:   userSvc,
		Auth:   authSvc,
		Email:  emailSvc,
		Verify: verifySvc,
		Follow: followSvc,
	}
}
