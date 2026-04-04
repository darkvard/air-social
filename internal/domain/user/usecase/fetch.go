package usecase

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type FetchDeps struct {
	Repo  user.Repository
	Cache user.Cache
	Link  common.LinkProvider
}

type fetchUseCase struct {
	deps FetchDeps
}

func NewFetchUseCase(d FetchDeps) *fetchUseCase {
	return &fetchUseCase{deps: d}
}

func (u *fetchUseCase) GetByID(ctx context.Context, id int64) (*user.User, error) {
	usr, err := u.deps.Repo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return usr, nil
}

func (u *fetchUseCase) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	usr, err := u.deps.Repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return usr, nil
}

func (u *fetchUseCase) GetSummary(ctx context.Context, id int64) (*user.UserSummary, error) {
	return u.deps.Cache.Get(ctx, id, func(ctx context.Context) (*user.UserSummary, error) {
		d, err := u.deps.Repo.GetByID(ctx, id)
		if err != nil {
			return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
		}
		return u.toUserSummary(d), nil
	})
}

func (u *fetchUseCase) toUserSummary(d *user.User) *user.UserSummary {
	return &user.UserSummary{
		ID:         d.ID,
		FullName:   d.Profile.FullName,
		Avatar:     u.deps.Link.PublicFile(d.Profile.Avatar),
		CoverImage: u.deps.Link.PublicFile(d.Profile.CoverImage),
		Verified:   d.Status.Verified,
	}
}
