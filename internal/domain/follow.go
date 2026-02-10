package domain

import (
	"context"
	"time"
)

const (
	SortLatest   = "latest"
	SortOldest   = "oldest"
	SortNameASC  = "name_asc"
	SortNameDESC = "name_desc"
)

type FollowRepository interface {
	Create(ctx context.Context, follow *Follow) error
	Delete(ctx context.Context, followerID, followeeID int64) error

	CountFollowings(ctx context.Context, userID int64) (int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)

	GetFollowers(ctx context.Context, params FollowParams) ([]User, error)
	GetFollowings(ctx context.Context, params FollowParams) ([]User, error)

	IsFollowing(ctx context.Context, userID int64, targetIDs []int64) (map[int64]bool, error)
	IsFollowedBy(ctx context.Context, userID int64, targetIDs []int64) (map[int64]bool, error)
}

type Follow struct {
	FollowerID int64
	FolloweeID int64
	CreatedAt  time.Time
}
type FollowParams struct {
	QueryParams
	TargetUserID  int64
	CurrentUserID int64
}

type SocialUser struct {
	User     User
	Relation Relationship
}

type Relationship struct {
	IsFollowing  bool // Me -> Them
	IsFollowedBy bool // Them -> Me
}
