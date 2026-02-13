package service

import (
	"context"

	"air-social/internal/domain"
)

type PostService interface {
	CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error)
	GetPostDetail(ctx context.Context, postID int64) (*domain.Post, error)
	GetUserPosts(ctx context.Context, userID int64, cursor int64, limit int) ([]domain.Post, error)
	UpdatePost(ctx context.Context, params domain.UpdatePostParams) (*domain.Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
}

type PostServiceImpl struct {
	postRepo domain.PostRepository
}

func NewPostService(postRepo domain.PostRepository) *PostServiceImpl {
	return &PostServiceImpl{postRepo: postRepo}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, params domain.CreatePostParams) (*domain.Post, error) {
	return nil, nil
}

func (s *PostServiceImpl) GetPostDetail(ctx context.Context, postID int64) (*domain.Post, error) {
	return nil, nil
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
