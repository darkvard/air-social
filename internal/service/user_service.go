package service

import (
	"context"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserService interface {
	CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	VerifyEmail(ctx context.Context, email string) error
}

type UserServiceImpl struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{repo: repo}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error) {
	if existing, _ := s.repo.GetByEmail(ctx, in.Email); existing != nil {
		return nil, pkg.ErrAlreadyExists
	}

	u := &domain.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: in.PasswordHash,
		Profile:      nil,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &domain.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Profile:   u.Profile,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, email string) error {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.Verified = true
	return s.repo.Update(ctx, user)
}
