package provider

import (
	"air-social/internal/domain/auth"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/media"
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
	cache shared.Cache,
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
	deps := auth.Deps{
		TokenRepo:     nil,
		TokenProvider: nil,
		UserFetch:     nil,
		UserAccount:   nil,
		Cache:         nil,
	}
	return auth.NewUseCase(deps)
}

func NewMediaUseCase() media.UseCase {
	deps := media.Deps{
		Bucket:  media.Bucket{},
		Storage: nil,
		Link:    nil,
		Route:   nil,
	}
	return media.NewUseCase(deps)

}

func NewFollowUseCase() follow.Usecase {
	deps := follow.Deps{
		FollowRepo:  nil,
		UserFetcher: nil,
	}
	return follow.NewUseCase(deps)
}
