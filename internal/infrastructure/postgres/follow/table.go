package follow

import (
	"time"

	"air-social/internal/domain/follow"
)

// FollowerID: the person who performs the action (subscriber)
// FolloweeID: the person being followed (start)
type Table struct {
	FollowerID int64     `db:"follower_id"`
	FolloweeID int64     `db:"followee_id"`
	CreatedAt  time.Time `db:"created_at"`
}

func (m *Table) ToDomain() *follow.Follow {
	return &follow.Follow{
		FollowerID: m.FollowerID,
		FolloweeID: m.FolloweeID,
		CreatedAt:  m.CreatedAt,
	}
}

func FromDomain(f *follow.Follow) *Table {
	return &Table{
		FollowerID: f.FollowerID,
		FolloweeID: f.FolloweeID,
		CreatedAt:  f.CreatedAt,
	}
}
