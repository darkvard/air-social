package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"air-social/internal/cache"
	"air-social/internal/domain"
	mess "air-social/internal/infrastructure/messaging"
	"air-social/internal/routes"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenInfo, error)
	Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.TokenInfo, error)
	Logout(ctx context.Context, req *domain.LogoutRequest) error
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest, isValidateReturn bool) error
}

type AuthServiceImpl struct {
	users     UserService
	tokens    TokenService
	publisher domain.EventPublisher
	routes    routes.Registry
	cache     cache.CacheStorage
}

func NewAuthService(
	us UserService,
	ts TokenService,
	pub domain.EventPublisher,
	rr routes.Registry,
	cs cache.CacheStorage,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		users:     us,
		tokens:    ts,
		publisher: pub,
		routes:    rr,
		cache:     cs,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, error) {
	hashedPwd, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.users.CreateUser(ctx, &domain.CreateUserInput{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPwd,
	})
	if err != nil {
		return nil, err
	}

	eventID := uuid.NewString()
	ttl := cache.ThirtyMinutesTime
	token := uuid.NewString()
	if err := s.storeEmailVerification(ctx, token, user.Email, ttl); err != nil {
		pkg.Log().Errorw("failed to store verification email", "error", err, "event", eventID, "email", user.Email)
	}
	if err := s.publisher.Publish(
		ctx, mess.EmailVerifyQueueConfig.RoutingKey,
		domain.EventPayload{
			EventID:   eventID,
			EventType: domain.EmailVerify,
			Timestamp: time.Now().UTC(),
			Data: domain.EventEmailData{
				Email:  user.Email,
				Name:   user.Username,
				Link:   s.routes.VerifyEmailURL(token),
				Expiry: pkg.FormatTTLVerbose(ttl),
			},
		},
	); err != nil {
		pkg.Log().Errorw("failed to send verification email", "error", err, "event", eventID, "email", user.Email)
	}

	return user, nil
}

// NOTE: bcrypt only processes the first 72 bytes of the input.
// Any bytes beyond that are silently ignored.
// => Two different passwords may generate the same hash if their first 72 bytes are identical.
//
// To avoid this issue:
// - Limit password length to 72 bytes (add input validation), OR
// - Pre-hash the password using SHA-256 before passing it to bcrypt (any -> 32 bytes)
func hashPassword(plainText string) (string, error) {
	// Pre-hash using SHA-256 to avoid 72-byte limitation in bcrypt
	sha := sha256.Sum256([]byte(plainText))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}

func (s *AuthServiceImpl) storeEmailVerification(ctx context.Context, token, email string, ttl time.Duration) error {
	return s.cache.Set(ctx, cache.GetEmailVerificationKey(token), email, ttl)
}

func (s *AuthServiceImpl) getEmailVerification(ctx context.Context, token string) (string, error) {
	var email string
	if err := s.cache.Get(ctx, cache.GetEmailVerificationKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	email, err := s.getEmailVerification(ctx, token)
	if err != nil {
		return err
	}

	return s.users.VerifyEmail(ctx, email)
}

func (s *AuthServiceImpl) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *domain.TokenInfo, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, pkg.ErrInvalidCredentials
	}

	if !verifyPassword(req.Password, user.PasswordHash) {
		return nil, nil, pkg.ErrInvalidCredentials
	}

	tokens, err := s.tokens.CreateSession(ctx, user.ID, req.DeviceID)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func verifyPassword(plainPassword, hashPassword string) bool {
	sha := sha256.Sum256([]byte(plainPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), sha[:])
	return err == nil
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

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	token := uuid.NewString()
	ttl := cache.FifteenMinutesTime
	if err := s.storeEmailResetPassword(ctx, user.Email, token, ttl); err != nil {
		return err
	}

	payload := domain.EventPayload{
		EventID:   uuid.NewString(),
		EventType: domain.EmailResetPassword,
		Timestamp: time.Now().UTC(),
		Data: domain.EventEmailData{
			Email:  user.Email,
			Name:   user.Username,
			Link:   s.routes.ResetPasswordURL(token),
			Expiry: pkg.FormatTTLVerbose(ttl),
		},
	}

	return s.publisher.Publish(ctx, mess.EmailResetPasswordQueueConfig.RoutingKey, payload)
}

func (s *AuthServiceImpl) getEmailResetPassword(ctx context.Context, token string) (string, error) {
	var email string
	if err := s.cache.Get(ctx, cache.GetEmailResetPasswordKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (s *AuthServiceImpl) storeEmailResetPassword(ctx context.Context, email, token string, ttl time.Duration) error {
	return s.cache.Set(ctx, cache.GetEmailResetPasswordKey(token), email, ttl)
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest, isValidateReturn bool) error {
	email, err := s.getEmailResetPassword(ctx, req.Token)
	if err != nil {
		return err
	}
	if isValidateReturn {
		return nil
	}

	hashedPwd, err := hashPassword(req.Password)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, email, hashedPwd)
}
