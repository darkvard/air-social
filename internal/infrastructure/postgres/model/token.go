package model

import (
	"time"

	"air-social/internal/domain"
)

type RefreshToken struct {
	ID        int64      `db:"id"`
	UserID    int64      `db:"user_id"`
	DeviceID  string     `db:"device_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func (m *RefreshToken) ToDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		DeviceID:  m.DeviceID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}

func FromDomainRefreshToken(t *domain.RefreshToken) *RefreshToken {
	return &RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		DeviceID:  t.DeviceID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		CreatedAt: t.CreatedAt,
	}
}
