package feed

import (
	"context"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
)

type QueryUseCase interface {
	GetNewsfeed(ctx context.Context, viewerID int64, params common.CursorQueryParams[int64]) (common.CursorPaginatedResult[*post.Post, int64], error)
}

type CommandUseCase interface {
	DistributePost(ctx context.Context, postID int64, authorID int64, timestamp int64) error
	RevokePost(ctx context.Context, postID int64, authorID int64) error
}

type UseCase struct {
	Query   QueryUseCase
	Command CommandUseCase
}
