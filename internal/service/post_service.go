package service

import (
	"context"
	"strings"

	"air-social/internal/domain"
	"air-social/pkg"
)

type PostService interface {
	CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error)
	GetPostDetail(ctx context.Context, id int64) (*domain.Post, error)
	GetUserPosts(ctx context.Context, userID int64, cursor int64, limit int) ([]domain.Post, error)
	UpdatePost(ctx context.Context, params domain.UpdatePostParams) (*domain.Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
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

func (s *PostServiceImpl) GetUserPosts(ctx context.Context, userID int64, cursor int64, limit int) ([]domain.Post, error) {
	return nil, nil
}

func (s *PostServiceImpl) UpdatePost(ctx context.Context, params domain.UpdatePostParams) (*domain.Post, error) {
	return nil, nil
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, postID int64, userID int64) error {
	return nil
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
