package usecase

import "context"

type PresenceStore interface {
	IsOnlineBatch(ctx context.Context, userIDs []int64) (map[int64]bool, error)
}

type PresenceDeps struct {
	Store PresenceStore
}

type PresenceUseCase struct {
	deps PresenceDeps
}

func NewPresenceUseCase(deps PresenceDeps) *PresenceUseCase {
	return &PresenceUseCase{deps: deps}
}

func (u *PresenceUseCase) GetBatchStatus(ctx context.Context, userIDs []int64) (map[int64]bool, error) {
	return u.deps.Store.IsOnlineBatch(ctx, userIDs)
}
