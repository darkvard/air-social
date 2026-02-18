package provider

import (
	"air-social/internal/domain"
	"air-social/internal/domain/auth/token"
	userdomain "air-social/internal/domain/user"
	"air-social/internal/infrastructure/postgres/repository"
	userinfra "air-social/internal/infrastructure/postgres/user"
)

type Repositories struct {
	User   domain.UserRepository
	Token  domain.TokenRepository
	Follow domain.FollowRepository
	Post   domain.PostRepository

	TokenProvider token.Provider
}

func NewRepositories(infra *Infrastructures) *Repositories {
	return &Repositories{
		User:   repository.NewUserRepository(infra.DB),
		Token:  repository.NewTokenRepository(infra.DB),
		Follow: repository.NewFollowRepository(infra.DB),
		Post:   repository.NewPostRepository(infra.DB),
	}
}

// todo: replace Repositories
type Repository struct {
	User userdomain.Repository
}


func NewRepository(infra *Infrastructures) *Repository {
	return &Repository{
		User: userinfra.NewRepository(infra.DB),
	}
}