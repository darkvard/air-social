package service

import (
	"context"

	"air-social/internal/domain/user"
	"air-social/pkg"
)

type UserService struct {
	repo user.Repository
}

func NewUserService(repo user.Repository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, in *user.CreateUserInput) (*user.UserResponse, error) {
	if existing, _ := s.repo.GetByEmail(ctx, in.Email); existing != nil {
		return nil, pkg.ErrAlreadyExists
	}

	u := &user.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: in.PasswordHash,
		Profile:      nil,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return &user.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Profile:   u.Profile,
		CreatedAt: u.CreatedAt,
	}, nil
}
