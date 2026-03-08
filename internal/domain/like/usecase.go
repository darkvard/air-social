package like

import (
	"context"

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
	inserted, ownerID, err := u.repo.InsertPostLike(ctx, postID, userID)
	if err != nil {
		switch err {
		case pkg.ErrNotFound:
			return nil
		case pkg.ErrInvalidData:
			return pkg.NewError(err, "post not found")
		default:
			return pkg.OrInternalError(err)
		}
	}

	if inserted {
		_ = u.addLikePostEvent(ctx, true, postID, userID, ownerID)
	}
	return nil
}

func (u *usecase) UnlikePost(ctx context.Context, postID, userID int64) error {
	err := u.repo.DeletePostLike(ctx, postID, userID)
	if err != nil {
		return pkg.NewError(err, "delete post like failed")
	}

	_ = u.addLikePostEvent(ctx, false, postID, userID, -1)
	return nil
}

func (u *usecase) LikeComment(ctx context.Context, commentID, userID int64) error {
	inserted, ownerID, err := u.repo.InsertCommentLike(ctx, commentID, userID)
	if err != nil {
		switch err {
		case pkg.ErrNotFound:
			return nil
		case pkg.ErrInvalidData:
			return pkg.NewError(err, "comment not found")
		default:
			return pkg.OrInternalError(err)
		}
	}

	if inserted {
		_ = u.addLikeCommentEvent(ctx, true, commentID, userID, ownerID)
	}

	return nil
}

func (u *usecase) UnlikeComment(ctx context.Context, commentID, userID int64) error {
	err := u.repo.DeleteCommentLike(ctx, commentID, userID)
	if err != nil {
		return pkg.NewError(err, "delete comment like failed")
	}

	_ = u.addLikeCommentEvent(ctx, false, commentID, userID, -1)

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

// todo thieu owner
func (u *usecase) addLikePostEvent(ctx context.Context, isLike bool, postID, userID, ownerID int64) error {
	typ := common.EventPostLike
	data := common.LikeEventPayload{
		TargetID: postID,
		ActorID:  userID,
		OwnerID:  ownerID,
		IsLiked:  isLike,
		Typ:      typ,
	}
	return u.event.Publish(ctx, common.NewEvent(typ, data))
}

func (u *usecase) addLikeCommentEvent(ctx context.Context, isLike bool, commentID, userID, ownerID int64) error {
	typ := common.EventCommentLike
	data := common.LikeEventPayload{
		TargetID: commentID,
		ActorID:  userID,
		OwnerID:  ownerID,
		IsLiked:  isLike,
		Typ:      typ,
	}
	return u.event.Publish(ctx, common.NewEvent(typ, data))
}
