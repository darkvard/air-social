package stats

import (
	"time"

	"air-social/internal/domain/stats"
)

type PostTable struct {
	PostID        int64     `db:"post_id"`
	LikesCount    int32     `db:"likes_count"`
	CommentsCount int32     `db:"comments_count"`
	SharesCount   int32     `db:"shares_count"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (m *PostTable) ToDomain() *stats.PostStats {
	return &stats.PostStats{
		PostID:        m.PostID,
		LikesCount:    m.LikesCount,
		CommentsCount: m.CommentsCount,
		SharesCount:   m.SharesCount,
		UpdatedAt:     m.UpdatedAt,
	}
}

func FromDomainPostStat(d *stats.PostStats) *PostTable {
	if d == nil {
		return nil
	}
	return &PostTable{
		PostID:        d.PostID,
		LikesCount:    d.LikesCount,
		CommentsCount: d.CommentsCount,
		SharesCount:   d.SharesCount,
		UpdatedAt:     d.UpdatedAt,
	}
}
