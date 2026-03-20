package follow

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, followerID, followeeID int64) error
	Delete(ctx context.Context, followerID, followeeID int64) error

	CountFollowings(ctx context.Context, userID int64) (int64, error)
	CountFollowers(ctx context.Context, userID int64) (int64, error)

	GetFollowers(ctx context.Context, params GetFollowsParams) ([]FollowUser, int64, error)
	GetFollowings(ctx context.Context, params GetFollowsParams) ([]FollowUser, int64, error)

	GetRelationship(ctx context.Context, userID, targetID int64) (Relationship, error)
	GetRelationships(ctx context.Context, userID int64, targetIDs []int64) (map[int64]Relationship, error)

	GetFollowerIDs(ctx context.Context, userID int64) ([]int64, error)
}
