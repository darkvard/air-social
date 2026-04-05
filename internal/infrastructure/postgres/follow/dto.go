package follow

import "air-social/internal/domain/follow"

type RelationshipRow struct {
	IsFollowing bool `db:"is_following"`
	IsFollower  bool `db:"is_follower"`
}

func (r RelationshipRow) ToDomain() follow.Relationship {
	return follow.Relationship{
		IsFollowing: r.IsFollowing,
		IsFollower:  r.IsFollower,
	}
}