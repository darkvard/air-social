package stats

import "context"

type Repository interface {
	BulkUpsertPostStats(ctx context.Context, params PostParams) error

	BulkUpsertCommentStats(ctx context.Context, params CommentParams) error

	GetPostsStats(ctx context.Context, postIDs []int64) ([]PostStats, error)

	GetCommentsStats(ctx context.Context, commentIDs []int64) ([]CommentStats, error)

	ReconcilePostStats(ctx context.Context, postIDs []int64) error

	ReconcileCommentStats(ctx context.Context, commentIDs []int64) error
}
