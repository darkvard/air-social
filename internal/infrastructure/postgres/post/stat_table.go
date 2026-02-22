package post

import (
	"time"

	"air-social/internal/domain/post"
)

type StatTable struct {
	PostID        int64     `db:"post_id"`
	LikesCount    int32     `db:"likes_count"`
	CommentsCount int32     `db:"comments_count"`
	SharesCount   int32     `db:"shares_count"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (m *StatTable) ToDomain() *post.Stat {
	return &post.Stat{
		PostID:        m.PostID,
		LikesCount:    m.LikesCount,
		CommentsCount: m.CommentsCount,
		SharesCount:   m.SharesCount,
		UpdatedAt:     m.UpdatedAt,
	}
}

func FromDomainStat(d *post.Stat) *StatTable {
	if d == nil {
		return nil
	}
	return &StatTable{
		PostID:        d.PostID,
		LikesCount:    d.LikesCount,
		CommentsCount: d.CommentsCount,
		SharesCount:   d.SharesCount,
		UpdatedAt:     d.UpdatedAt,
	}
}
