package follow

import (
	"time"
)

type Follow struct {
	FollowerID int64
	FolloweeID int64
	CreatedAt  time.Time
}

type FollowUser struct {
	ID         int64
	FullName   string
	Avatar     string
	IsVerified bool
	Relationship
}

type Relationship struct {
	IsFollowing bool // i am follow them
	IsFollower  bool // they are following me
}
