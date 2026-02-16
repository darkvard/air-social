package domain

import (
	"context"
	"time"
)

type LikeRepository interface {
	// todo db transaction: Insert Like + Update LikesCount at Post table
	LikePost(ctx context.Context, postID, userID int64) error
	UnlikePost(ctx context.Context, postID, userID int64) error

	// todo db transaction: Insert Like + Update LikesCount at Comment table
	LikeComment(ctx context.Context, commentID, userID int64) error
	UnlikeComment(ctx context.Context, commentID, userID int64) error
	
	// todo: check status: map to dto isLiked
	IsPostLiked(ctx context.Context, postID, userID int64) (bool, error)
	IsCommentLiked(ctx context.Context, commentID, userID int64) (bool, error)
}

type PostLike struct {
	PostID    int64
	UserID    int64
	CreatedAt time.Time
}

type CommentLike struct {
	CommentID int64
	UserID    int64
	CreatedAt time.Time
}