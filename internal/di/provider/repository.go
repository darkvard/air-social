package provider

import (
	tokendomain "air-social/internal/domain/auth/token"
	followdomain "air-social/internal/domain/follow"
	postdomain "air-social/internal/domain/post"
	userdomain "air-social/internal/domain/user"
	followinfra "air-social/internal/infrastructure/postgres/follow"
	postinfra "air-social/internal/infrastructure/postgres/post"
	tokeninfra "air-social/internal/infrastructure/postgres/token"
	userinfra "air-social/internal/infrastructure/postgres/user"
)

type Repository struct {
	User   userdomain.Repository
	Token  tokendomain.Repository
	Follow followdomain.Repository
	Post   postdomain.Repository
}

func NewRepository(infra *Infrastructure) Repository {
	return Repository{
		User:   userinfra.NewRepository(infra.DB),
		Token:  tokeninfra.NewRepository(infra.DB),
		Follow: followinfra.NewRepository(infra.DB),
		Post:   postinfra.NewRepository(infra.DB),
	}
}
