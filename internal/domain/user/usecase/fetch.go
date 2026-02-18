package usecase

import (
	"context"

	"air-social/internal/domain/shared"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type fetchUseCase struct {
	repo  user.Repository
	cache shared.CacheStorage
	link  shared.AppLinkProvider
}

func NewFetchUseCase(d user.Deps) *fetchUseCase {
	return &fetchUseCase{
		repo:  d.Repo,
		cache: d.Cache,
		link:  d.Link,
	}
}

func (u *fetchUseCase) GetByID(ctx context.Context, id int64) (*user.User, error) {
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return user, nil
}

func (u *fetchUseCase) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return user, nil

}

func (u *fetchUseCase) GetSummary(ctx context.Context, id int64) (*user.UserSummaryResult, error) {
	cached, err := getUserCache(ctx, u.cache, id)
	if err == nil && cached != nil {
		return cached, nil
	}

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	summary := u.toUserSummary(user)
	_ = setUserCache(ctx, u.cache, summary)

	return summary, nil
}

func (u *fetchUseCase) toUserSummary(d *user.User) *user.UserSummaryResult {
	return &user.UserSummaryResult{
		ID:         d.ID,
		FullName:   d.Profile.FullName,
		Avatar:     u.link.PublicFile(d.Profile.Avatar),
		CoverImage: u.link.PublicFile(d.Profile.CoverImage),
		Verified:   d.Status.Verified,
	}
}
