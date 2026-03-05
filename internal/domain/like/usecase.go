package like

import (
	"context"
	"errors"
	"fmt"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type UseCase interface {
	LikePost(ctx context.Context, postID, userID int64) error
	UnlikePost(ctx context.Context, postID, userID int64) error
	LikeComment(ctx context.Context, commentID, userID int64) error
	UnlikeComment(ctx context.Context, commentID, userID int64) error

	IsPostLiked(ctx context.Context, postIDs []int64, userID int64) (map[int64]bool, error)
	IsCommentLiked(ctx context.Context, commentIDs []int64, userID int64) (map[int64]bool, error)
}

type Deps struct {
	Repo  Repository
	Event common.EventPublisher
}

type usecase struct {
	repo  Repository
	event common.EventPublisher
}

func NewUsecase(deps Deps) UseCase {
	return &usecase{
		repo:  deps.Repo,
		event: deps.Event,
	}
}

func (u *usecase) LikePost(ctx context.Context, postID, userID int64) error {
	inserted, err := u.repo.InsertPostLike(ctx, postID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidData) {
			return pkg.NewError(err, "post not found")
		}
		return pkg.OrInternalError(err)
	}

	if inserted {
		// todo push notification & update stats
		_ = u.event.Publish(ctx, common.Event{})
	}

	return nil
}

func (u *usecase) IsPostLiked(ctx context.Context, postIDs []int64, userID int64) (map[int64]bool, error) {
	if len(postIDs) == 0 {
		return map[int64]bool{}, nil
	}

	likedIDs, err := u.repo.GetPostLiked(ctx, postIDs, userID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}

	result := make(map[int64]bool, len(postIDs))
	for _, id := range postIDs {
		result[id] = false
	}
	for _, id := range likedIDs {
		result[id] = true
	}

	return result, nil
}

func (u *usecase) IsCommentLiked(ctx context.Context, commentIDs []int64, userID int64) (map[int64]bool, error) {
	if len(commentIDs) == 0 {
		return map[int64]bool{}, nil
	}

	likedIDs, err := u.repo.GetCommentLiked(ctx, commentIDs, userID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}

	result := make(map[int64]bool, len(commentIDs))
	for _, id := range commentIDs {
		result[id] = false
	}
	for _, id := range likedIDs {
		result[id] = true
	}

	return result, nil
}

func (u *usecase) UnlikePost(ctx context.Context, postID, userID int64) error {
	err := u.repo.DeletePostLike(ctx, postID, userID)
	if err != nil {
		return fmt.Errorf("delete post like failed: %w", err)
	}

	// todo update stats
	_ = u.event.Publish(ctx, common.Event{})

	return nil
}

func (u *usecase) LikeComment(ctx context.Context, commentID, userID int64) error {
	inserted, err := u.repo.InsertCommentLike(ctx, commentID, userID)
	if err != nil {
		if errors.Is(err, pkg.ErrInvalidData) {
			return pkg.NewError(err, "comment not found")
		}
		return pkg.OrInternalError(err)
	}

	if inserted {
		// todo push notification & update stats
		_ = u.event.Publish(ctx, common.Event{})
	}

	return nil
}

func (u *usecase) UnlikeComment(ctx context.Context, commentID, userID int64) error {
	err := u.repo.DeleteCommentLike(ctx, commentID, userID)
	if err != nil {
		return fmt.Errorf("delete comment like failed: %w", err)
	}

	// todo update stats
	_ = u.event.Publish(ctx, common.Event{})

	return nil
}
