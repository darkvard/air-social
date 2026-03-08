package like

import "context"

type Repository interface {
	InsertPostLike(ctx context.Context, postID, userID int64) (bool, int64, error)
	DeletePostLike(ctx context.Context, postID, userID int64) error

	InsertCommentLike(ctx context.Context, commentID, userID int64) (bool, int64, error)
	DeleteCommentLike(ctx context.Context, commentID, userID int64) error

	GetPostLiked(ctx context.Context, postIDs []int64, userID int64) ([]int64, error)
	GetCommentLiked(ctx context.Context, commentIDs []int64, userID int64) ([]int64, error)
}
