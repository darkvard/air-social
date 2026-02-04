package domain

import (
	"context"
	"time"
)

type FollowRepository interface {
	Create(ctx context.Context, follow *Follow) error
	Delete(ctx context.Context, followerID, followeeID int64) error

	CountFollowings(ctx context.Context, userID int64) (int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)

	IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error)
	GetFollowings(ctx context.Context, params FollowParams) ([]User, error)
	GetFollowers(ctx context.Context, params FollowParams) ([]User, error)
}

type Follow struct {
	FollowerID int64
	FolloweeID int64
	CreatedAt  time.Time
}

type FollowResult struct {
	Users []User
	Total int64
	Page  int
	Limit int
}

type FollowParams struct {
	UserID int64
	Page   int
	Limit  int
}

func (f FollowParams) GetPage() int {
	if f.Page <= 0 {
		return 1
	}
	return f.Page
}

func (f FollowParams) GetLimit() int {
	if f.Limit <= 0 {
		return 10
	}
	if f.Limit > 100 {
		return 100
	}
	return f.Limit
}

func (f FollowParams) GetOffset() int {
	return (f.GetPage() - 1) * f.GetLimit()
}
