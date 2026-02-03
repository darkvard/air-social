package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"air-social/internal/domain"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, input domain.RegisterParams) (*domain.User, error)
	Logout(ctx context.Context, input domain.LogoutParams) error
	Login(ctx context.Context, input domain.LoginParams) (*domain.User, domain.TokenInfo, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, input domain.ResetPasswordParams) error
	IsResetPasswordTokenValid(ctx context.Context, token string) bool
	RefreshToken(ctx context.Context, refreshToken string) (domain.TokenInfo, error)
	VerifyEmail(ctx context.Context, emailToken string) error
	GetPublicURL(key string) string
}

type AuthServiceImpl struct {
	userSvc   UserService
	tokenSvc  TokenService
	verifySvc VerifyService
	cache     domain.CacheStorage
}

func NewAuthService(userSvc UserService, tokenSvc TokenService, verifySvc VerifyService, cache domain.CacheStorage) *AuthServiceImpl {
	return &AuthServiceImpl{
		userSvc:   userSvc,
		tokenSvc:  tokenSvc,
		verifySvc: verifySvc,
		cache:     cache,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, input domain.RegisterParams) (*domain.User, error) {
	var empty *domain.User

	passwordHashed, err := hashPassword(input.Password)
	if err != nil {
		return empty, pkg.ErrInternal
	}
	params := domain.CreateUserParams{
		Email:          input.Email,
		Username:       input.Username,
		PasswordHashed: passwordHashed,
	}

	user, err := s.userSvc.CreateUser(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrAlreadyExists)
	}

	if err := s.verifySvc.SendEmailVerification(ctx, user.Email, user.Username); err != nil {
		pkg.Log().Errorw("failed to send verification email", "error", err, "email", user.Email)
	}
	return user, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, input domain.LogoutParams) error {
	if s.isBlockedAccessToken(ctx, input.Token.AccessToken) {
		return pkg.ErrUnauthorized
	}

	var err error
	if input.IsAllDevices {
		err = s.tokenSvc.RevokeAllUserSessions(ctx, input.UserID)
	} else {
		err = s.tokenSvc.RevokeDeviceSession(ctx, input.UserID, input.DeviceID)
	}

	if err == nil {
		s.blockAccessToken(ctx, input.Token)
	}
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) Login(ctx context.Context, input domain.LoginParams) (*domain.User, domain.TokenInfo, error) {
	var emptyUser *domain.User
	var emptyToken domain.TokenInfo

	user, err := s.userSvc.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return emptyUser, emptyToken, pkg.ErrInvalidCredentials
		}
		return emptyUser, emptyToken, err
	}

	if !verifyPassword(input.Password, user.PasswordHash) {
		return emptyUser, emptyToken, pkg.ErrInvalidCredentials
	}

	tokens, err := s.tokenSvc.CreateSession(ctx, user.ID, input.DeviceID)
	if err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err)
	}

	return user, tokens, nil
}

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	return s.verifySvc.SendPasswordReset(ctx, user.Email, user.Username)
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, input domain.ResetPasswordParams) error {
	email, err := s.verifySvc.VerifyPasswordResetToken(ctx, input.EmailToken)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	passwordHashed, err := hashPassword(input.Password)
	if err != nil {
		return pkg.ErrInternal
	}

	err = s.userSvc.UpdatePassword(ctx, email, passwordHashed)
	if err != nil {
		return pkg.OrInternalError(err)
	}

	_ = s.verifySvc.InvalidatePasswordResetToken(ctx, input.EmailToken)
	return nil
}

func (s *AuthServiceImpl) IsResetPasswordTokenValid(ctx context.Context, emailToken string) bool {
	_, err := s.verifySvc.VerifyPasswordResetToken(ctx, emailToken)
	if err != nil {
		return false
	}
	return true
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, emailToken string) error {
	email, err := s.verifySvc.VerifyEmailToken(ctx, emailToken)
	if err != nil {
		return pkg.ErrBadRequest
	}

	err = s.userSvc.VerifyEmail(ctx, email)
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (domain.TokenInfo, error) {
	var empty domain.TokenInfo

	tokens, err := s.tokenSvc.Refresh(ctx, refreshToken)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrUnauthorized)
	}
	return tokens, nil
}

func (s *AuthServiceImpl) GetPublicURL(key string) string {
	return s.userSvc.GetPublicURL(key)
}

// Internal helpers

// hashPassword generates a bcrypt hash of the password using the default cost.
//
// To circumvent bcrypt's 72-byte input truncation limit, the password is
// pre-hashed using SHA-256 before being passed to bcrypt. This ensures
// passwords of any length are securely handled.
func hashPassword(plainText string) (string, error) {
	// SHA-256 produces a fixed 32-byte hash, safe for bcrypt.
	sha := sha256.Sum256([]byte(plainText))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(plainPassword, hashPassword string) bool {
	sha := sha256.Sum256([]byte(plainPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), sha[:])
	return err == nil
}

func (s *AuthServiceImpl) isBlockedAccessToken(ctx context.Context, token string) bool {
	key := domain.GetBlacklistTokenKey(token)
	exists, _ := s.cache.IsExist(ctx, key)
	return exists
}

func (s *AuthServiceImpl) blockAccessToken(ctx context.Context, token domain.TokenMeta) {
	ttl := time.Until(time.Unix(token.ExpiresAt, 0))
	if ttl > 0 {
		key := domain.GetBlacklistTokenKey(token.AccessToken)
		_ = s.cache.Set(ctx, key, "revoked", ttl)
	}
}
