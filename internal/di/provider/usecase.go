package provider

import (
	"air-social/internal/domain"
	"air-social/internal/domain/user"
	uuc "air-social/internal/domain/user/usecase"
)

// todo: replace Service
type UseCase struct {
	User user.UseCases
}

func NewUseCase() *UseCase {
	return &UseCase{
		 
	}
}

func NewUserUseCases(
	repo user.Repository,
	cache domain.CacheStorage,
	url domain.URLFactory,
) user.UseCases {
	d := user.Deps{
		Repo:  repo,
		Cache: cache,
		URL:   url,
	}
	return user.UseCases{
		Profile: uuc.NewProfileUseCase(d),
		Account: uuc.NewAccountUseCase(d),
		Fetch:   uuc.NewFetchUseCase(d),
	}
}
