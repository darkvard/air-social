package like

import "context"

type Repository interface {
	InsertPostLike(ctx context.Context, postID, userID int64) (bool, int64, error)
	DeletePostLike(ctx context.Context, postID, userID int64) (int64, error) // returns post owner ID

	InsertCommentLike(ctx context.Context, commentID, userID int64) (bool, int64, error)
	DeleteCommentLike(ctx context.Context, commentID, userID int64) (int64, error) // returns comment owner ID

	GetPostLiked(ctx context.Context, postIDs []int64, userID int64) ([]int64, error)
	GetCommentLiked(ctx context.Context, commentIDs []int64, userID int64) ([]int64, error)

	GetPostLikers(ctx context.Context, params GetCursorParams) ([]Liker, error)
	GetCommentLikers(ctx context.Context, params GetCursorParams) ([]Liker, error)
}
