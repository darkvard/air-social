package user

import (
	"context"

	"air-social/internal/domain/media"
	"air-social/internal/domain/shared"
)

type AccountUseCase interface {
	Authenticate(ctx context.Context, params AuthenticateParams) (*User, error)
	CreateUser(ctx context.Context, params CreateParams) (*User, error)
	ChangePassword(ctx context.Context, params ChangePasswordParams) error
	ResetPassword(ctx context.Context, params ResetPasswordParams) error
	VerifyEmail(ctx context.Context, email string) error
}

type ProfileUseCase interface {
	UpdateProfile(ctx context.Context, params UpdateParams) (*User, error)
	UpdateAvatar(ctx context.Context, params media.ConfirmFileParams) error
	UpdateCover(ctx context.Context, params media.ConfirmFileParams) error
}

type FetchUseCase interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetSummary(ctx context.Context, id int64) (*UserSummaryResult, error)
}

type UseCases struct {
	Account AccountUseCase
	Profile ProfileUseCase
	Fetch   FetchUseCase
}

type Deps struct {
	Repo  Repository
	Cache shared.CacheStorage
	Link  shared.AppLinkProvider
	Media media.UseCase
}
