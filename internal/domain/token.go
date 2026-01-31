package domain

import (
	"context"
	"time"
)

type TokenRepository interface {
	Create(ctx context.Context, t RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (RefreshToken, error)
	UpdateRevoked(ctx context.Context, id int64) error
	UpdateRevokedByUser(ctx context.Context, userID int64) error
	UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error
	DeleteExpiredAndRevoked(ctx context.Context, expiredBefore time.Time, revokedBefore time.Time) error
}

const AuditRetentionPeriod = 30 * 24 * time.Hour

type RefreshToken struct {
	ID        int64      `db:"id"`
	UserID    int64      `db:"user_id"`
	DeviceID  string     `db:"device_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type TokenInfo struct {
	Type            string    `json:"type"`
	AccessToken     string    `json:"access_token"`
	RefreshToken    string    `json:"refresh_token"`
	AccessExpireAt  time.Time `json:"access_expire_at"`
	RefreshExpireAt time.Time `json:"refresh_expire_at"`
}
