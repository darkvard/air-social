package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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
	ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error
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

	input := &domain.CreateUserInput{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPwd,
	}

	user, err := s.users.CreateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.sendEmailVerify(ctx, user.Email, user.Username); err != nil {
		pkg.Log().Errorw("failed to send verification email", "error", err, "user_id", user.ID, "email", user.Email)
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

func (s *AuthServiceImpl) sendEmailVerify(ctx context.Context, email, name string) error {
	token := uuid.NewString()
	key := getEmailVerificationKey(token)
	ttl := 24 * time.Hour

	if err := s.cache.Set(ctx, key, email, ttl); err != nil {
		return err
	}

	verifyData := domain.EventEmailData{
		Email:  email,
		Name:   name,
		Link:   s.routes.VerifyEmailURL(token),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}

	payload := domain.EventPayload{
		EventID:   uuid.NewString(),
		EventType: domain.EmailVerify,
		Timestamp: time.Now(),
		Data:      verifyData,
	}

	return s.publisher.Publish(ctx, mess.EmailVerifyQueueConfig.RoutingKey, payload)
}

func getEmailVerificationKey(token string) string {
	return fmt.Sprintf(cache.WorkerEmailVerify+"%s", token)
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	key := getEmailVerificationKey(token)
	var email string
	if err := s.cache.Get(ctx, key, &email); err != nil || email == "" {
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

	return &domain.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Profile:   user.Profile,
		CreatedAt: user.CreatedAt,
	}, tokens, nil
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
	key := getEmailResetPasswordKey(token)
	ttl := 15 * time.Minute
	if err := s.cache.Set(ctx, key, user.Email, ttl); err != nil {
		return err
	}

	resetData := domain.EventEmailData{
		Email:  user.Email,
		Name:   user.Username,
		Link:   s.routes.ResetPasswordURL(token),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}

	payload := domain.EventPayload{
		EventID:   uuid.NewString(),
		EventType: domain.EmailResetPassword,
		Timestamp: time.Now(),
		Data:      resetData,
	}

	return s.publisher.Publish(ctx, mess.EmailResetPasswordQueueConfig.RoutingKey, payload)
}

func getEmailResetPasswordKey(token string) string {
	return fmt.Sprintf(cache.WorkerEmailReset+"%s", token)
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	// check redis
	return nil
}
