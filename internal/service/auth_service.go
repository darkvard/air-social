package service

import (
	"context"

	"air-social/internal/domain/auth"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type AuthService struct {
	users *UserService
	jwt   pkg.JWTAuth
	hash  pkg.Hasher
}

func NewAuthService(users *UserService, jwt pkg.JWTAuth, hash pkg.Hasher) *AuthService {
	return &AuthService{
		users: users,
		jwt:   jwt,
		hash:  hash,
	}
}

func (s *AuthService) Register(ctx context.Context, req *auth.RegisterRequest) (*user.UserResponse, error) {
	hashedPwd, err := s.hash.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	input := &user.CreateUserInput{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPwd,
	}

	return s.users.CreateUser(ctx, input)
}
