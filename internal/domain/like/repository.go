package like

import "context"

type Repository interface {
	InsertPostLike(ctx context.Context, postID, userID int64) (bool, error)
	DeletePostLike(ctx context.Context, postID, userID int64) error

	InsertCommentLike(ctx context.Context, commentID, userID int64) (bool, error)
	DeleteCommentLike(ctx context.Context, commentID, userID int64) error
}
