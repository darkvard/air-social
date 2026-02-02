package di

import (
	"air-social/internal/domain"
	"air-social/internal/infrastructure/postgres/repository"
)

type Repositories struct {
	User  domain.UserRepository
	Token domain.TokenRepository
}

func initRepositories(infra *Infrastructures) *Repositories {
	return &Repositories{
		User:  repository.NewUserRepository(infra.DB),
		Token: repository.NewTokenRepository(infra.DB),
	}
}
