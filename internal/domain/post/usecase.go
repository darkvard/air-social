package post

import (
	"context"
	"errors"
	"strings"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type UseCase interface {
	GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error)
	CreatePost(ctx context.Context, params CreateParams) (*Post, error)
	UpdatePost(ctx context.Context, params UpdateParams) (*Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
	GetUserPosts(ctx context.Context, params GetCursorParams) (common.CursorPaginatedResult[Post, int64], error)
}

type MediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type Deps struct {
	PostRepo      Repository
	MediaVerifier MediaVerifier
	Event         common.EventPublisher
}

type usecase struct {
	postRepo      Repository
	mediaVerifier MediaVerifier
	event         common.EventPublisher
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		postRepo:      deps.PostRepo,
		mediaVerifier: deps.MediaVerifier,
		event:         deps.Event,
	}
}

func (u *usecase) GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error) {
	post, err := u.postRepo.GetDetail(ctx, postID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(err, "post not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	// todo: them usecase moi cho stat
	// todo map isLiked, count.....

	return post, nil
}

func (u *usecase) CreatePost(ctx context.Context, params CreateParams) (*Post, error) {
	if strings.TrimSpace(params.Content) == "" && len(params.Media) == 0 {
		return nil, pkg.NewError(pkg.ErrInvalidData, "content or media is required")
	}

	post := &Post{
		OriginalPostID: params.OriginalPostID,
		UserID:         params.UserID,
		Content:        params.Content,
		Visibility:     params.Visibility,
	}
	if len(params.Media) > 0 {
		media, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
			}
			return nil, pkg.OrInternalError(err)
		}
		post.Media = media
	}

	if err := u.postRepo.Create(ctx, post); err != nil {
		if errors.Is(err, pkg.ErrInvalidData) {
			return nil, pkg.NewError(pkg.ErrInvalidData, "original post not found")
		}
		return nil, err
	}

	if post.OriginalPostID != nil {
		_ = u.addShareEvent(ctx, *post, true)
	}

	return post, nil
}

func (u *usecase) UpdatePost(ctx context.Context, params UpdateParams) (*Post, error) {
	post, err := u.getPostOwner(ctx, params.PostID, params.UserID)
	if err != nil {
		return nil, err
	}

	if params.Content != nil {
		post.Content = *params.Content
	}
	if params.Visibility != nil {
		post.Visibility = Visibility(*params.Visibility)
	}
	if params.Media != nil {
		media, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.NewError(pkg.ErrBadRequest, "media items not found or invalid")
			}
			return nil, pkg.OrInternalError(err)
		}
		post.Media = media // update
	} else {
		post.Media = nil // skip
	}

	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, pkg.ErrInternal
	}
	return post, nil
}

// DeletePost deletes a post by ID.
// If the post was a share, it publishes an EventPostShare (isShared=false).
func (u *usecase) DeletePost(ctx context.Context, postID, userID int64) error {
	post, err := u.getPostOwner(ctx, postID, userID)
	if err != nil {
		return err
	}
	if post.OriginalPostID != nil {
		_ = u.addShareEvent(ctx, *post, false)
	}
	return pkg.OrInternalError(u.postRepo.Delete(ctx, postID))
}

func (u *usecase) GetUserPosts(ctx context.Context, params GetCursorParams) (common.CursorPaginatedResult[Post, int64], error) {
	params.Query.NormalizePagination()

	var empty common.CursorPaginatedResult[Post, int64]

	posts, err := u.postRepo.GetUserPosts(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	// todo map isLiked, count.....

	result := common.NewCursorPaginatedResult(posts, params.Query.Limit)
	return result, nil
}

func (u *usecase) validateMedia(ctx context.Context, params []MediaParams) ([]Media, error) {
	size := len(params)

	keys := make([]string, size)
	for i, m := range params {
		keys[i] = m.MediaKey
	}
	if err := u.mediaVerifier.VerifyMedia(ctx, keys); err != nil {
		return nil, err
	}

	media := make([]Media, size)
	for i, v := range params {
		media[i] = v.ToDomain()
	}
	return media, nil
}

func (u *usecase) getPostOwner(ctx context.Context, postID, viewerID int64) (*Post, error) {
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(err, "post not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	if post.UserID != viewerID {
		return nil, pkg.ErrForbidden
	}
	return post, nil
}

func (u *usecase) addShareEvent(ctx context.Context, post Post, isShare bool) error {
	data := common.ShareEventPayload{
		OriginalPostID: *post.OriginalPostID,
		NewPostID:      post.ID,
		ActorID:        post.UserID,
		IsShared:       isShare,
	}
	return u.event.Publish(ctx, common.NewEvent(common.EventPostShare, data))
}
