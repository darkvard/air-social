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

const AuditRetentionPeriod = 30 * 24 * time.Hour

type RefreshToken struct {
	ID        int64
	UserID    int64
	DeviceID  string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type TokenInfo struct {
	Type            string
	AccessToken     string
	RefreshToken    string
	AccessExpireAt  time.Time
	RefreshExpireAt time.Time
}
