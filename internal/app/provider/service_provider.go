package provider

import (
	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/routes"
	"air-social/internal/service"
)

type ServiceProvider struct {
	User  service.UserService
	Auth  service.AuthService
	Token service.TokenService
}

func NewServiceProvider(
	repo *RepoProvider,
	cfg config.TokenConfig,
	pub domain.EventPublisher,
	rr routes.Registry,
	cs cache.CacheStorage,
) *ServiceProvider {
	token := service.NewTokenService(repo.Token, cfg)
	user := service.NewUserService(repo.User)
	auth := service.NewAuthService(user, token, pub, rr, cs)
	return &ServiceProvider{
		User:  user,
		Auth:  auth,
		Token: token,
	}
}
