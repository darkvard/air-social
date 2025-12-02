package service

import (
	"context"

	"air-social/internal/domain"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenInfo, error)
}

type AuthServiceImpl struct {
	users  UserService
	tokens TokenService
	hash   pkg.Hasher
}

func NewAuthService(users UserService, tokens TokenService, hash pkg.Hasher) *AuthServiceImpl {
	return &AuthServiceImpl{
		users:  users,
		tokens: tokens,
		hash:   hash,
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

func (s *AuthServiceImpl) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenInfo, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, pkg.ErrInvalidCredentials
	}

	if !s.hash.Verify(req.Password, user.PasswordHash) {
		return nil, nil, pkg.ErrInvalidCredentials
	}

	tokens, err := s.tokens.GenerateTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	return &domain.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Profile:   user.Profile,
		CreatedAt: user.CreatedAt,
	}, tokens, nil
}
