package service

import (
	"context"
	"errors"

	"air-social/internal/domain"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenInfo, error)
	Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.TokenInfo, error)
	Logout(ctx context.Context, req *domain.LogoutRequest) error
}

type AuthServiceImpl struct {
	users  UserService
	tokens TokenService
	hash   pkg.Hasher
	queue  domain.EventQueue
}

func NewAuthService(users UserService, tokens TokenService, hash pkg.Hasher, queue domain.EventQueue) *AuthServiceImpl {
	return &AuthServiceImpl{
		users:  users,
		tokens: tokens,
		hash:   hash,
		queue:  queue,
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

	tokens, err := s.tokens.CreateSession(ctx, user.ID, req.DeviceID)
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

func (s *AuthServiceImpl) Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.TokenInfo, error) {
	tokens, err := s.tokens.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, pkg.ErrTokenExpired),
			errors.Is(err, pkg.ErrTokenRevoked),
			errors.Is(err, pkg.ErrNotFound):
			return nil, pkg.ErrUnauthorized
		default:
			return nil, pkg.ErrInternal
		}
	}
	return tokens, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, req *domain.LogoutRequest) error {
	if req.IsAllDevices {
		return s.tokens.RevokeAllUserSessions(ctx, req.UserID)
	} else {
		return s.tokens.RevokeDeviceSession(ctx, req.UserID, req.DeviceID)

	}
}
