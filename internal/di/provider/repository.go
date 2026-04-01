package provider

import (
	tokendomain "air-social/internal/domain/auth/token"
	commentdomain "air-social/internal/domain/comment"
	followdomain "air-social/internal/domain/follow"
	likedomain "air-social/internal/domain/like"
	postdomain "air-social/internal/domain/post"
	searchdomain "air-social/internal/domain/search"
	statsdomain "air-social/internal/domain/stats"
	userdomain "air-social/internal/domain/user"
	commentinfra "air-social/internal/infrastructure/postgres/comment"
	followinfra "air-social/internal/infrastructure/postgres/follow"
	likeinfra "air-social/internal/infrastructure/postgres/like"
	postinfra "air-social/internal/infrastructure/postgres/post"
	searchinfra "air-social/internal/infrastructure/postgres/search"
	statsinfra "air-social/internal/infrastructure/postgres/stats"
	tokeninfra "air-social/internal/infrastructure/postgres/token"
	userinfra "air-social/internal/infrastructure/postgres/user"
)

type Repository struct {
	User    userdomain.Repository
	Token   tokendomain.Repository
	Follow  followdomain.Repository
	Post    postdomain.Repository
	Like    likedomain.Repository
	Comment commentdomain.Repository
	Stats   statsdomain.Repository
	Search  searchdomain.Repository
}

func NewRepository(infra *Infrastructure) Repository {
	return Repository{
		User:    userinfra.NewRepository(infra.Postgres),
		Token:   tokeninfra.NewRepository(infra.Postgres),
		Follow:  followinfra.NewRepository(infra.Postgres),
		Post:    postinfra.NewRepository(infra.Postgres),
		Like:    likeinfra.NewRepository(infra.Postgres),
		Comment: commentinfra.NewRepository(infra.Postgres),
		Stats:   statsinfra.NewRepository(infra.Postgres),
		Search:  searchinfra.NewRepository(infra.Postgres),
	}
}
