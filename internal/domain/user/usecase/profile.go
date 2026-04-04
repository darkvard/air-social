package usecase

import (
	"context"

	"air-social/internal/domain/media"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type MediaConfirmer interface {
	ConfirmUpload(ctx context.Context, params []media.ConfirmParams) ([]string, error)
}

type ProfileDeps struct {
	Repo  user.Repository
	Cache user.Cache
	Media MediaConfirmer
}

type profileUseCase struct {
	deps ProfileDeps
}

func NewProfileUseCase(d ProfileDeps) *profileUseCase {
	return &profileUseCase{deps: d}
}

func (u *profileUseCase) UpdateProfile(ctx context.Context, params user.UpdateParams) (*user.User, error) {
	var empty *user.User

	user, err := u.deps.Repo.GetByID(ctx, params.UserID)
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

	if err := u.deps.Repo.UpdateProfile(ctx, user); err != nil {
		return empty, pkg.OrInternalError(err)
	}
	_ = u.deps.Cache.Invalidate(ctx, user.ID)

	return user, nil
}

func (u *profileUseCase) UpdateAvatar(ctx context.Context, params media.ConfirmParams) (*user.User, error) {
	if err := u.updateImageUrl(ctx, params, u.deps.Repo.UpdateAvatar); err != nil {
		return nil, err
	}

	result, err := u.deps.Repo.GetByID(ctx, params.EntityID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *profileUseCase) UpdateCover(ctx context.Context, params media.ConfirmParams) (*user.User, error) {
	if err := u.updateImageUrl(ctx, params, u.deps.Repo.UpdateCover); err != nil {
		return nil, err
	}

	result, err := u.deps.Repo.GetByID(ctx, params.EntityID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *profileUseCase) updateImageUrl(
	ctx context.Context,
	params media.ConfirmParams,
	updateFunc func(context.Context, int64, string) error,
) error {
	if !media.IsDomainFeatureValid(params) {
		return pkg.ErrBadRequest
	}

	keys, err := u.deps.Media.ConfirmUpload(ctx, []media.ConfirmParams{params})
	if err != nil || len(keys) == 0 {
		return pkg.OrInternalError(err, pkg.ErrBadRequest, pkg.ErrNotFound)
	}

	if err := updateFunc(ctx, params.EntityID, keys[0]); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = u.deps.Cache.Invalidate(ctx, params.EntityID)
	return nil
}
