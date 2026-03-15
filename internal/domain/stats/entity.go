package stats

import "time"

const (
	SystemName   = "social"
	FeatureStats = "stats"
)

type PostStats struct {
	PostID        int64
	LikesCount    int32
	CommentsCount int32
	SharesCount   int32
	UpdatedAt     time.Time
}

type CommentStats struct {
	CommentID    int64
	LikesCount   int32
	RepliesCount int32
	UpdatedAt    time.Time
}
