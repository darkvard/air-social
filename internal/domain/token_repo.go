package domain

import (
	"context"
	"time"
)

type TokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	UpdateRevoked(ctx context.Context, id int64) error
	UpdateRevokedByUser(ctx context.Context, userID int64) error
	UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error
	DeleteExpiredAndRevoked(ctx context.Context, expiredBefore time.Time, revokedBefore time.Time) error
}
