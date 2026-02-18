package token

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	UpdateRevoked(ctx context.Context, tokenID int64) error
	UpdateRevokedByUser(ctx context.Context, userID int64) error
	UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error
}
