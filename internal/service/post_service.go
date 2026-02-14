package service

import (
	"context"
	"strings"

	"air-social/internal/domain"
	"air-social/pkg"
)

type PostService interface {
	CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error)
	UpdatePost(ctx context.Context, params domain.UpdatePostParams) (*domain.Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
	GetPostDetail(ctx context.Context, id int64) (*domain.Post, error)
	GetUserPosts(ctx context.Context, userID int64, param domain.CursorQueryParams) (domain.CursorPaginatedResult[domain.Post], error)
}

type PostServiceImpl struct {
	postRepo domain.PostRepository
	mediaSvc MediaService
	userSvc  UserService
}

func NewPostService(postRepo domain.PostRepository, mediaSvc MediaService, userSvc UserService) *PostServiceImpl {
	return &PostServiceImpl{
		postRepo: postRepo,
		mediaSvc: mediaSvc,
		userSvc:  userSvc,
	}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error) {
	user, err := s.userSvc.GetSummary(ctx, params.UserID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	if strings.TrimSpace(params.Content) == "" && len(params.Media) == 0 {
		return nil, pkg.ErrInvalidData
	}

	post := &domain.Post{
		UserID:     params.UserID,
		Content:    params.Content,
		Visibility: params.Visibility,
		User:       user,
	}

	if len(params.Media) > 0 {
		media, err := s.validateMedia(ctx, params.Media)
		if err != nil {
			return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
		}
		post.Media = media
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostServiceImpl) GetPostDetail(ctx context.Context, id int64) (*domain.Post, error) {
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	user, err := s.userSvc.GetSummary(ctx, post.UserID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	post.User = user

	return post, nil
}

func (s *PostServiceImpl) UpdatePost(ctx context.Context, params domain.UpdatePostParams) (*domain.Post, error) {
	isOwner, err := s.postRepo.IsOwner(ctx, params.PostID, params.UserID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	if !isOwner {
		return nil, pkg.ErrForbidden
	}

	post, err := s.postRepo.GetByID(ctx, params.PostID)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	if params.Content != nil {
		post.Content = *params.Content
	}
	if params.Visibility != nil {
		post.Visibility = domain.PostVisibility(*params.Visibility)
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	userSummary, _ := s.userSvc.GetSummary(ctx, post.UserID)
	post.User = userSummary

	return post, nil
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, postID int64, userID int64) error {
	isOwner, err := s.postRepo.IsOwner(ctx, postID, userID)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	if !isOwner {
		return pkg.ErrForbidden
	}

	err = s.postRepo.Delete(ctx, postID)

	return pkg.OrInternalError(err)
}

func (s *PostServiceImpl) GetUserPosts(ctx context.Context, userID int64, param domain.CursorQueryParams) (domain.CursorPaginatedResult[domain.Post], error) {
	var empty domain.CursorPaginatedResult[domain.Post]
	param.EnsureDefaults()

	summary, err := s.userSvc.GetSummary(ctx, userID)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	posts, err := s.postRepo.GetUserPosts(ctx, userID, param)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	return s.assemblePaginatedResult(posts, summary, param.Limit), nil
}

// Internal helpers

func (s *PostServiceImpl) assemblePaginatedResult(posts []domain.Post, user *domain.UserSummary, limit int) domain.CursorPaginatedResult[domain.Post] {
	if len(posts) == 0 {
		return domain.NewCursorPaginatedResult(posts, 0, false)
	}

	// Handle "Limit + 1" logic to determine Next Page
	hasNextPage := false
	if len(posts) > limit {
		hasNextPage = true
		posts = posts[:limit]
	}

	// Calculate Next Cursor
	var nextCursor int64 = 0
	if hasNextPage {
		nextCursor = posts[len(posts)-1].ID
	}

	// Attach User Profile to each Post
	for i := range posts {
		posts[i].User = user
	}

	return domain.NewCursorPaginatedResult(posts, nextCursor, hasNextPage)
}

func (s *PostServiceImpl) validateMedia(ctx context.Context, params []domain.PostMediaParams) ([]domain.PostMedia, error) {
	size := len(params)

	keys := make([]string, size)
	for i, m := range params {
		keys[i] = m.MediaKey
	}
	if err := s.mediaSvc.VerifyMedia(ctx, keys); err != nil {
		return nil, err
	}

	media := make([]domain.PostMedia, size)
	for i, v := range params {
		media[i] = s.toPostMedia(v)
	}

	return media, nil
}

func (s *PostServiceImpl) toPostMedia(param domain.PostMediaParams) domain.PostMedia {
	return domain.PostMedia{
		MediaKey:  param.MediaKey,
		MediaType: param.MediaType,
		Metadata: domain.PostMediaMetadata{
			Width:    param.Width,
			Height:   param.Height,
			Duration: param.Duration,
			Size:     param.Size,
			FileName: param.FileName,
		},
	}
}
