package comment

import (
	"context"
	"errors"

	"air-social/internal/domain/follow"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type UseCase interface {
	GetComments(ctx context.Context, postID int64, params GetCursorParams) ([]Comment, int64, error)

	GetReplies(ctx context.Context, parentID int64, params GetCursorParams) ([]Comment, int64, error)

	CreateComment(ctx context.Context, params CreateParams) (*Comment, error)

	UpdateComment(ctx context.Context, params UpdateParams) (*Comment, error)

	DeleteComment(ctx context.Context, commentID, userID int64) error
}

type PostFetcher interface {
	GetPostDetail(ctx context.Context, postID, viewerID int64) (*post.Post, error)
}

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type MediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type Deps struct {
	CommentRepo   Repository
	PostFetcher   PostFetcher
	FollowChecker FollowChecker
	MediaVerifier MediaVerifier
}

type usecase struct {
	commentRepo   Repository
	postFetcher   PostFetcher
	followChecker FollowChecker
	mediaVerifier MediaVerifier
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		commentRepo:   deps.CommentRepo,
		postFetcher:   deps.PostFetcher,
		followChecker: deps.FollowChecker,
		mediaVerifier: deps.MediaVerifier,
	}
}

func (u *usecase) GetComments(ctx context.Context, postID int64, params GetCursorParams) ([]Comment, int64, error) {
	return nil, -1, nil
}

func (u *usecase) GetReplies(ctx context.Context, parentID int64, params GetCursorParams) ([]Comment, int64, error) {
	return nil, -1, nil
}

func (u *usecase) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	comment, err := u.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.NewError(pkg.ErrBadRequest, "comment not found")
		}
		return pkg.OrInternalError(err)
	}

	if comment.UserID != userID {
		return pkg.NewError(pkg.ErrForbidden, "only the author has that right")
	}

	if err := u.commentRepo.Delete(ctx, comment.ID); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (u *usecase) UpdateComment(ctx context.Context, params UpdateParams) (*Comment, error) {
	comment, err := u.commentRepo.GetByID(ctx, params.CommentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "comment not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	if comment.UserID != params.UserID {
		return nil, pkg.NewError(pkg.ErrForbidden, "only the author has that right")
	}

	if params.Content != nil {
		comment.Content = *params.Content
	}
	if params.Media != nil {
		res, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			return nil, err
		}
		comment.Media = res
	}

	if err := u.commentRepo.Update(ctx, comment); err != nil {
		return nil, pkg.OrInternalError(err)
	}
	return comment, nil
}

func (u *usecase) CreateComment(ctx context.Context, params CreateParams) (*Comment, error) {
	_, err := u.validatePostVisibility(ctx, params.PostID, params.UserID)
	if err != nil {
		return nil, err
	}

	finalParentID, err := u.resolveParentID(ctx, params.ParentID, params.PostID)
	if err != nil {
		return nil, err
	}

	if err := u.mediaVerifier.VerifyMedia(ctx, params.GetMediaKeys()); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
		}
		return nil, pkg.OrInternalError(err)
	}

	comment := &Comment{
		PostID:   params.PostID,
		UserID:   params.UserID,
		ParentID: finalParentID,
		Content:  params.Content,
		Media:    params.Media,
	}

	if err := u.commentRepo.Create(ctx, comment); err != nil {
		return nil, pkg.OrInternalError(err)
	}
	return comment, nil
}

func (u *usecase) validatePostVisibility(ctx context.Context, postID, userID int64) (*post.Post, error) {
	p, err := u.postFetcher.GetPostDetail(ctx, postID, userID)
	if err != nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "post not found")
	}
	if p == nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "post not found")
	}

	if p.UserID == userID {
		return p, nil
	}

	switch p.Visibility {
	case post.VisibilityPrivate:
		return nil, pkg.NewError(pkg.ErrForbidden, "post is private")
	case post.VisibilityFollowers:
		rel, err := u.followChecker.GetRelationship(ctx, userID, p.UserID)
		if err != nil || !rel.IsFollowing {
			return nil, pkg.NewError(pkg.ErrForbidden, "only followers can comment")
		}
	}
	return p, nil
}

func (u *usecase) resolveParentID(ctx context.Context, parentID *int64, postID int64) (*int64, error) {
	if parentID == nil || *parentID <= 0 {
		return nil, nil
	}

	parent, err := u.commentRepo.GetByID(ctx, *parentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "parent comment not found or has been deleted")
		}
		return nil, pkg.OrInternalError(err)
	}

	if parent.PostID != postID {
		return nil, pkg.NewError(pkg.ErrBadRequest, "parent comment belongs to another post")
	}

	// One-level depth
	if parent.ParentID != nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "cannot reply to a sub-comment, please reply to the root comment")
	}

	return parentID, nil
}

func (u *usecase) validateMedia(ctx context.Context, params []Media) ([]Media, error) {
	size := len(params)
	keys := make([]string, size)
	for i, m := range params {
		keys[i] = m.MediaKey
	}

	if err := u.mediaVerifier.VerifyMedia(ctx, keys); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
		}
		return nil, pkg.OrInternalError(err)
	}

	return params, nil
}
