package stats

import (
	"context"

	"air-social/internal/domain/stats/cache"
)

type UseCase interface {
	SyncPostStats(ctx context.Context) error
	SyncCommentStats(ctx context.Context) error
}

type Deps struct {
	Repo  Repository
	Cache cache.Provider
}

type usecase struct {
	repo Repository
	cache cache.Provider
}

func NewUseCase(deps Deps) UseCase {
	return &usecase{
		repo: deps.Repo,
		cache: deps.Cache,
	}
}

func (u *usecase) SyncPostStats(ctx context.Context) error {
	// get data from cache
	// bulk upsert to db
	// clear cache
	return nil
}

func (u *usecase) SyncCommentStats(ctx context.Context) error {
	return nil
}
