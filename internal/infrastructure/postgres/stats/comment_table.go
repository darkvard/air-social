package stats

import (
	"time"

	"air-social/internal/domain/stats"
)

type CommentTable struct {
	CommentID    int64     `db:"comment_id"`
	LikesCount   int32     `db:"likes_count"`
	RepliesCount int32     `db:"replies_count"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m *CommentTable) ToDomain() *stats.CommentStats {
	return &stats.CommentStats{
		CommentID:    m.CommentID,
		LikesCount:   m.LikesCount,
		RepliesCount: m.RepliesCount,
		UpdatedAt:    m.UpdatedAt,
	}
}

func FromDomainCommentStat(d *stats.CommentStats) *CommentTable {
	if d == nil {
		return nil
	}
	return &CommentTable{
		CommentID:    d.CommentID,
		LikesCount:   d.LikesCount,
		RepliesCount: d.RepliesCount,
		UpdatedAt:    d.UpdatedAt,
	}
}
