package model

import (
	"time"

	"air-social/internal/domain"
)
// todo: remove

type Follow struct {
	FollowerID int64     `db:"follower_id"`
	FolloweeID int64     `db:"followee_id"`
	CreatedAt  time.Time `db:"created_at"`
}

func (m *Follow) ToDomain() *domain.Follow {
	return &domain.Follow{
		FollowerID: m.FollowerID,
		FolloweeID: m.FolloweeID,
		CreatedAt:  m.CreatedAt,
	}
}

func FromDomainFollow(f *domain.Follow) *Follow {
	return &Follow{
		FollowerID: f.FollowerID,
		FolloweeID: f.FolloweeID,
		CreatedAt:  f.CreatedAt,
	}
}
