package token

import "time"

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

type TokenClaims struct {
	UserID   int64
	DeviceID string
	Role     int64
}
