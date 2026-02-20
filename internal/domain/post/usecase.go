package post

import (
	"context"
	"strings"

	"air-social/internal/domain/shared"
	"air-social/pkg"
)

// todo: redesign database: post, like, comment
// => fix hot row contention
type UseCase interface {
	GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error)
	CreatePost(ctx context.Context, params CreateParams) (*Post, error)
	UpdatePost(ctx context.Context, params UpdateParams) (*Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
	GetUserPosts(ctx context.Context, params GetCursorParams) (shared.CursorPaginatedResult[Post, int64], error)
}

type mediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type Deps struct {
	PostRepo      Repository
	MediaVerifier mediaVerifier
}

type usecase struct {
	postRepo      Repository
	mediaVerifier mediaVerifier
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		postRepo:      deps.PostRepo,
		mediaVerifier: deps.MediaVerifier,
	}
}

func (u *usecase) GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error) {
	post, err := u.postRepo.GetDetail(ctx, postID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	// todo map isLiked, count.....

	return post, nil
}

func (u *usecase) CreatePost(ctx context.Context, params CreateParams) (*Post, error) {
	if strings.TrimSpace(params.Content) == "" && len(params.Media) == 0 {
		return nil, pkg.ErrInvalidData
	}

	post := &Post{
		UserID:     params.UserID,
		Content:    params.Content,
		Visibility: params.Visibility,
	}

	if len(params.Media) > 0 {
		media, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
		}
		post.Media = media
	}

	if err := u.postRepo.Create(ctx, post); err != nil {
		return nil, err
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

	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, pkg.ErrInternal
	}
	return post, nil
}

func (u *usecase) DeletePost(ctx context.Context, postID, userID int64) error {
	_, err := u.getPostOwner(ctx, postID, userID)
	if err != nil {
		return err
	}
	return pkg.OrInternalError(u.postRepo.Delete(ctx, postID))
}

func (u *usecase) GetUserPosts(ctx context.Context, params GetCursorParams) (shared.CursorPaginatedResult[Post, int64], error) {
	params.Query.NormalizePagination()

	var empty shared.CursorPaginatedResult[Post, int64]

	posts, err := u.postRepo.GetUserPosts(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	// todo map isLiked, count.....

	result := shared.NewCursorPaginatedResult(posts, params.Query.Limit)
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
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	if post.UserID != viewerID {
		return nil, pkg.ErrForbidden
	}
	return post, nil
}
