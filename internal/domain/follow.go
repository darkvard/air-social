package domain

import (
	"context"
	"time"
)

type FollowRepository interface {
	Create(ctx context.Context, follow *Follow) error
	Delete(ctx context.Context, followerID, followeeID int64) error

	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)

	GetFollowings(ctx context.Context, userID int64, limit, offset int) ([]User, error)
	GetFollowers(ctx context.Context, userID int64, limit, offset int) ([]User, error)

	CountFollowings(ctx context.Context, userID int64) (int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)
}


type Follow struct {
	FollowerID int64
	FolloweeID int64
	CreatedAt  time.Time
}
