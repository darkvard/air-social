package usecase

import (
	"context"

	"air-social/internal/domain/media"
	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type MediaConfirmer interface {
	ConfirmUpload(ctx context.Context, params []media.ConfirmParams) ([]string, error)
}

type profileUseCase struct {
	repo  user.Repository
	cache shared.Cache
	media MediaConfirmer
}

func NewProfileUseCase(d Deps) *profileUseCase {
	return &profileUseCase{
		repo:  d.Repo,
		cache: d.Cache,
		media: d.Media,
	}
}

func (u *profileUseCase) UpdateProfile(ctx context.Context, params user.UpdateParams) (*user.User, error) {
	var empty *user.User

	user, err := u.repo.GetByID(ctx, params.UserID)
	if err != nil {
		return empty, err
	}
	if params.FullName != nil {
		user.Profile.FullName = *params.FullName
	}
	if params.Bio != nil {
		user.Profile.Bio = *params.Bio
	}
	if params.Location != nil {
		user.Profile.Location = *params.Location
	}
	if params.Website != nil {
		user.Profile.Website = *params.Website
	}
	if params.Username != nil {
		user.Username = *params.Username
	}

	if err := u.repo.UpdateProfile(ctx, user); err != nil {
		return empty, pkg.OrInternalError(err)
	}
	_ = clearUserCache(ctx, u.cache, user.ID)

	return user, nil
}

func (u *profileUseCase) UpdateAvatar(ctx context.Context, params media.ConfirmParams) error {
	return u.updateImageUrl(ctx, params, u.repo.UpdateAvatar)
}

func (u *profileUseCase) UpdateCover(ctx context.Context, params media.ConfirmParams) error {
	return u.updateImageUrl(ctx, params, u.repo.UpdateCover)
}

func (u *profileUseCase) updateImageUrl(
	ctx context.Context,
	params media.ConfirmParams,
	updateFunc func(context.Context, int64, string) error,
) error {
	if !media.IsDomainFeatureValid(params) {
		return pkg.ErrBadRequest
	}

	keys, err := u.media.ConfirmUpload(ctx, []media.ConfirmParams{params})
	if err != nil || len(keys) == 0 {
		return pkg.OrInternalError(err, pkg.ErrBadRequest)
	}

	if err := updateFunc(ctx, params.EntityID, keys[0]); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = clearUserCache(ctx, u.cache, params.EntityID)
	return nil
}
