package service

import (
	"context"
	"time"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserService interface {
	CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	GetByID(ctx context.Context, id int64) (*domain.UserResponse, error)
	VerifyEmail(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, pwd string) error
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
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u.ToResponse(), nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (*domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, email string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	user.Verified = true
	user.VerifiedAt = &now
	return s.repo.Update(ctx, user)
}

func (s *UserServiceImpl) UpdatePassword(ctx context.Context, email, pwd string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.PasswordHash = pwd
	return s.repo.Update(ctx, user)
}
