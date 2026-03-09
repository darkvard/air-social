package stats

import "context"

type Repository interface {
	BulkUpsertPostStats(ctx context.Context, params PostParams) error

	BulkUpsertCommentStats(ctx context.Context, params CommentParams) error
}
