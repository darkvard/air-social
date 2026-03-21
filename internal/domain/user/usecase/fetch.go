package usecase

import (
	"context"

	appcache "air-social/internal/cache"
	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type fetchUseCase struct {
	repo  user.Repository
	cache appcache.TieredStore[*user.UserSummary]
	link  common.LinkProvider
}

func NewFetchUseCase(d Deps) *fetchUseCase {
	return &fetchUseCase{
		repo:  d.Repo,
		cache: d.Cache,
		link:  d.Link,
	}
}

func (u *fetchUseCase) GetByID(ctx context.Context, id int64) (*user.User, error) {
	usr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	_ = u.cache.Set(ctx, GetKey(id), u.toUserSummary(usr))
	return usr, nil
}

func (u *fetchUseCase) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	usr, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	_ = u.cache.Set(ctx, GetKey(usr.ID), u.toUserSummary(usr))
	return usr, nil
}

func (u *fetchUseCase) GetSummary(ctx context.Context, id int64) (*user.UserSummary, error) {
	return getUserCache(ctx, u.cache, id, func(ctx context.Context) (*user.UserSummary, error) {
		d, err := u.repo.GetByID(ctx, id)
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
		Avatar:     u.link.PublicFile(d.Profile.Avatar),
		CoverImage: u.link.PublicFile(d.Profile.CoverImage),
		Verified:   d.Status.Verified,
	}
}
