package provider

import (
	"air-social/internal/service"
	"air-social/pkg"
)

type ServiceProvider struct {
	User service.UserService
	Auth service.AuthService
}

func NewServiceProvider(repo *RepoProvider, jwt pkg.JWTAuth, hash pkg.Hasher) *ServiceProvider {
	user := service.NewUserService(repo.User)
	auth := service.NewAuthService(user, jwt, hash)
	return &ServiceProvider{
		User: user,
		Auth: auth,
	}
}
