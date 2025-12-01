package domain

import (
	"context"
)

type TokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeAllByUser(ctx context.Context, userID int64) error
}
