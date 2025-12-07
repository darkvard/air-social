package provider

import (
	"air-social/internal/config"
	"air-social/internal/service"
	"air-social/pkg"
)

type ServiceProvider struct {
	User service.UserService
	Auth service.AuthService
}

func NewServiceProvider(repo *RepoProvider, cfg config.TokenConfig, hash pkg.Hasher) *ServiceProvider {
	token := service.NewTokenService(repo.Token, cfg, repo.Cache)
	user := service.NewUserService(repo.User)
	auth := service.NewAuthService(user, token, hash)
	return &ServiceProvider{
		User: user,
		Auth: auth,
	}
}
