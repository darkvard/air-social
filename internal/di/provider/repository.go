package provider

import (
	tokendomain "air-social/internal/domain/auth/token"
	followdomain "air-social/internal/domain/follow"
	likedomain "air-social/internal/domain/like"
	postdomain "air-social/internal/domain/post"
	userdomain "air-social/internal/domain/user"
	followinfra "air-social/internal/infrastructure/postgres/follow"
	likeinfra "air-social/internal/infrastructure/postgres/like"
	postinfra "air-social/internal/infrastructure/postgres/post"
	tokeninfra "air-social/internal/infrastructure/postgres/token"
	userinfra "air-social/internal/infrastructure/postgres/user"
)

type Repository struct {
	User   userdomain.Repository
	Token  tokendomain.Repository
	Follow followdomain.Repository
	Post   postdomain.Repository
	Like   likedomain.Repository
}

func NewRepository(infra *Infrastructure) Repository {
	return Repository{
		User:   userinfra.NewRepository(infra.DB),
		Token:  tokeninfra.NewRepository(infra.DB),
		Follow: followinfra.NewRepository(infra.DB),
		Post:   postinfra.NewRepository(infra.DB),
		Like:   likeinfra.NewRepository(infra.DB),
	}
}
