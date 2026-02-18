package provider

import (
	"air-social/internal/domain/auth"
	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
	uuc "air-social/internal/domain/user/usecase"
)

// todo: replace Service
type UseCase struct {
	User user.UseCases
	Auth auth.UseCase
}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func NewUserUseCases(
	repo user.Repository,
	cache shared.CacheStorage,
	links shared.AppLinkProvider,
) user.UseCases {
	deps := user.Deps{
		Repo:  repo,
		Cache: cache,
		Link:  links,
	}
	return user.UseCases{
		Profile: uuc.NewProfileUseCase(deps),
		Account: uuc.NewAccountUseCase(deps),
		Fetch:   uuc.NewFetchUseCase(deps),
	}
}

func NewAuthUseCase() auth.UseCase {
	d := auth.Deps{
		TokenRepository: nil,
		TokenProvider:   nil,
		UserFetch:       nil,
		UserAccount:     nil,
		Cache:           nil,
	}
	return auth.NewUseCase(d)
}
