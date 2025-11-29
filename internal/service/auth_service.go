package service

import (
	"context"

	"air-social/internal/domain"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
}

type AuthServiceImpl struct {
	users UserService
	jwt   pkg.JWTAuth
	hash  pkg.Hasher
}

func NewAuthService(users UserService, jwt pkg.JWTAuth, hash pkg.Hasher) *AuthServiceImpl {
	return &AuthServiceImpl{
		users: users,
		jwt:   jwt,
		hash:  hash,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	hashedPwd, err := s.hash.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	input := &domain.CreateUserInput{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPwd,
	}

	return s.users.CreateUser(ctx, input)
}
