package provider

import (
	td "air-social/internal/domain/auth/token"
	fd "air-social/internal/domain/follow"
	pd "air-social/internal/domain/post"
	ud "air-social/internal/domain/user"
	fi "air-social/internal/infrastructure/postgres/follow"
	pi "air-social/internal/infrastructure/postgres/post"
	ti "air-social/internal/infrastructure/postgres/token"
	ui "air-social/internal/infrastructure/postgres/user"
)

type Repository struct {
	User   ud.Repository
	Token  td.Repository
	Follow fd.Repository
	Post   pd.Repository
}

func NewRepository(infra *Infrastructure) Repository {
	return Repository{
		User:   ui.NewRepository(infra.DB),
		Token:  ti.NewRepository(infra.DB),
		Follow: fi.NewRepository(infra.DB),
		Post:   pi.NewRepository(infra.DB),
	}
}
