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
