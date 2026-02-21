package comment

import (
	"time"

	"air-social/internal/domain/comment"
)

type StatTable struct {
	CommentID    int64     `db:"comment_id"`
	LikesCount   int32     `db:"likes_count"`
	RepliesCount int32     `db:"replies_count"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m *StatTable) ToDomain() *comment.Stat {
	return &comment.Stat{
		CommentID:    m.CommentID,
		LikesCount:   m.LikesCount,
		RepliesCount: m.RepliesCount,
		UpdatedAt:    m.UpdatedAt,
	}
}

func FromDomainStat(d *comment.Stat) *StatTable {
	if d == nil {
		return nil
	}
	return &StatTable{
		CommentID:    d.CommentID,
		LikesCount:   d.LikesCount,
		RepliesCount: d.RepliesCount,
		UpdatedAt:    d.UpdatedAt,
	}
}
