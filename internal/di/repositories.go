package di

import (
	"air-social/internal/domain"
	"air-social/internal/infrastructure/postgres"
)

type Repositories struct {
	User  domain.UserRepository
	Token domain.TokenRepository
}

func initRepository(infra *Infrastructures) *Repositories {
	return &Repositories{
		User:  postgres.NewUserRepository(infra.DB),
		Token: postgres.NewTokenRepository(infra.DB),
	}
}
