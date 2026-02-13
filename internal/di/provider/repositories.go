package provider

import (
	"air-social/internal/domain"
	"air-social/internal/infrastructure/postgres/repository"
)

type Repositories struct {
	User   domain.UserRepository
	Token  domain.TokenRepository
	Follow domain.FollowRepository
	Post   domain.PostRepository
}

func NewRepositories(infra *Infrastructures) *Repositories {
	return &Repositories{
		User:   repository.NewUserRepository(infra.DB),
		Token:  repository.NewTokenRepository(infra.DB),
		Follow: repository.NewFollowRepository(infra.DB),
		Post:   repository.NewPostRepository(infra.DB),
	}
}
