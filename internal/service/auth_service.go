package service

import (
	"context"
	"errors"
	"time"

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
}

type UserAccountManager interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error)
	UpdatePassword(ctx context.Context, email, passwordHashed string) error
	VerifyEmail(ctx context.Context, email string) error
}

type SessionManager interface {
	CreateSession(ctx context.Context, userID int64, deviceID string) (domain.TokenInfo, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenInfo, error)
	RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error
}

type AuthServiceImpl struct {
	accountMgr UserAccountManager
	sessionMgr SessionManager
	verifySvc  VerifyService
	cache      domain.CacheStorage
}

func NewAuthService(accountMgr UserAccountManager, sessionMgr SessionManager, verifySvc VerifyService, cache domain.CacheStorage) *AuthServiceImpl {
	return &AuthServiceImpl{
		accountMgr: accountMgr,
		sessionMgr: sessionMgr,
		verifySvc:  verifySvc,
		cache:      cache,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, input domain.RegisterParams) (*domain.User, error) {
	var empty *domain.User

	passwordHashed, err := pkg.HashPassword(input.Password)
	if err != nil {
		return empty, pkg.ErrInternal
	}
	params := domain.CreateUserParams{
		Email:          input.Email,
		Username:       input.Username,
		PasswordHashed: passwordHashed,
	}

	user, err := s.accountMgr.CreateUser(ctx, params)
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
		err = s.sessionMgr.RevokeAllUserSessions(ctx, input.UserID)
	} else {
		err = s.sessionMgr.RevokeDeviceSession(ctx, input.UserID, input.DeviceID)
	}

	if err == nil {
		s.blockAccessToken(ctx, input.Token)
	}
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) Login(ctx context.Context, input domain.LoginParams) (*domain.User, domain.TokenInfo, error) {
	var emptyUser *domain.User
	var emptyToken domain.TokenInfo

	user, err := s.accountMgr.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return emptyUser, emptyToken, pkg.ErrInvalidCredentials
		}
		return emptyUser, emptyToken, err
	}

	if !pkg.VerifyPassword(input.Password, user.PasswordHash) {
		return emptyUser, emptyToken, pkg.ErrInvalidCredentials
	}

	tokens, err := s.sessionMgr.CreateSession(ctx, user.ID, input.DeviceID)
	if err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err)
	}

	return user, tokens, nil
}

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.accountMgr.GetByEmail(ctx, email)
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

	passwordHashed, err := pkg.HashPassword(input.Password)
	if err != nil {
		return pkg.ErrInternal
	}

	err = s.accountMgr.UpdatePassword(ctx, email, passwordHashed)
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

	err = s.accountMgr.VerifyEmail(ctx, email)
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (domain.TokenInfo, error) {
	var empty domain.TokenInfo

	tokens, err := s.sessionMgr.Refresh(ctx, refreshToken)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrUnauthorized)
	}
	return tokens, nil
}

// Internal helpers

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
