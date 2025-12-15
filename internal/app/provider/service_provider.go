package provider

import (
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/pkg"
)

type ServiceProvider struct {
	User  service.UserService
	Auth  service.AuthService
	Token service.TokenService
}

func NewServiceProvider(
	repo *RepoProvider,
	cfg config.TokenConfig,
	hash pkg.Hasher,
	queue domain.EventQueue,
) *ServiceProvider {
	token := service.NewTokenService(repo.Token, cfg)
	user := service.NewUserService(repo.User)
	auth := service.NewAuthService(user, token, hash, queue)
	return &ServiceProvider{
		User:  user,
		Auth:  auth,
		Token: token,
	}
}
