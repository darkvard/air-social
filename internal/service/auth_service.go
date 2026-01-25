package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"air-social/internal/domain"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/pkg"
)

type AuthService interface {
	Register(ctx context.Context, input domain.RegisterParams) (domain.UserResponse, error)
	Logout(ctx context.Context, input domain.LogoutParams) error
	Login(ctx context.Context, input domain.LoginParams) (domain.LoginResponse, error)

	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, input domain.ResetPasswordParams) error
	IsResetPasswordTokenValid(ctx context.Context, token string) bool

	RefreshToken(ctx context.Context, refreshToken string) (domain.TokenInfo, error)
	VerifyEmail(ctx context.Context, emailToken string) error
}

type AuthServiceImpl struct {
	userSvc  UserService
	tokenSvc TokenService
	url      domain.URLFactory
	cache    domain.CacheStorage
	event    domain.EventPublisher
}

func NewAuthService(userSvc UserService, tokenSvc TokenService, url domain.URLFactory, event domain.EventPublisher, cache domain.CacheStorage) *AuthServiceImpl {
	return &AuthServiceImpl{
		userSvc:  userSvc,
		tokenSvc: tokenSvc,
		url:      url,
		event:    event,
		cache:    cache,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, input domain.RegisterParams) (domain.UserResponse, error) {
	var empty domain.UserResponse

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

	s.sendEmailVerification(ctx, user.Email, empty.Username)
	return user, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, input domain.LogoutParams) error {
	var err error
	if input.IsAllDevices {
		err = s.tokenSvc.RevokeAllUserSessions(ctx, input.UserID)
	} else {
		err = s.tokenSvc.RevokeDeviceSession(ctx, input.UserID, input.DeviceID)
	}
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) Login(ctx context.Context, input domain.LoginParams) (domain.LoginResponse, error) {
	var empty domain.LoginResponse

	user, err := s.userSvc.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return empty, pkg.ErrInvalidCredentials
		}
		return empty, err
	}

	if !verifyPassword(input.Password, user.PasswordHash) {
		return empty, pkg.ErrInvalidCredentials
	}

	tokens, err := s.tokenSvc.CreateSession(ctx, user.ID, input.DeviceID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	userResponse := user.ToResponse()
	s.userSvc.ResolveMediaURLs(&userResponse)

	return domain.LoginResponse{User: userResponse, Token: tokens}, nil
}

func (s *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userSvc.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	s.sendEmailResetPassword(ctx, user.Email, user.Username)
	return nil
}

func (s *AuthServiceImpl) ResetPassword(ctx context.Context, input domain.ResetPasswordParams) error {
	email, err := s.getEmailResetPassword(ctx, input.EmailToken)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	passwordHashed, err := hashPassword(input.Password)
	if err != nil {
		return pkg.ErrInternal
	}

	err = s.userSvc.UpdatePassword(ctx, email, passwordHashed)
	return pkg.OrInternalError(err)
}

func (s *AuthServiceImpl) IsResetPasswordTokenValid(ctx context.Context, emailToken string) bool {
	email, err := s.getEmailResetPassword(ctx, emailToken)
	if err != nil {
		return false
	}
	if email == "" {
		return false
	}
	return true
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, emailToken string) error {
	email, err := s.getEmailVerification(ctx, emailToken)
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

// sendEmailVerification sends an email verification event to the event publisher.
func (s *AuthServiceImpl) sendEmailVerification(ctx context.Context, email, username string) {
	id := uuid.NewString()
	ttl := domain.ThirtyMinutesTime

	if err := s.storeEmailVerification(ctx, id, email, ttl); err != nil {
		pkg.Log().Errorw("[CACHE ERROR]", "from", "email_verification", "error", err)
		return
	}

	data := domain.EventEmailData{
		Email:  email,
		Name:   username,
		Link:   s.url.VerifyEmailLink(id),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}
	payload := domain.EventPayload{
		EventID:   id,
		EventType: domain.EmailVerify,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	if err := s.event.Publish(ctx, rabbitmq.EmailVerifyQueueConfig.RoutingKey, payload); err != nil {
		pkg.Log().Errorw("[EVENT QUEUE ERROR]", "from", "email_verification", "error", err)
	}
}

func (s *AuthServiceImpl) storeEmailVerification(ctx context.Context, token, email string, ttl time.Duration) error {
	return s.cache.Set(ctx, domain.GetEmailVerificationKey(token), email, ttl)
}

func (s *AuthServiceImpl) getEmailVerification(ctx context.Context, token string) (string, error) {
	var email string
	key := domain.GetEmailVerificationKey(token)
	if err := s.cache.Get(ctx, key, &email); err != nil {
		return "", err
	}
	return email, nil
}

// sendEmailResetPassword sends an email reset password event to the event publisher.
func (s *AuthServiceImpl) sendEmailResetPassword(ctx context.Context, email, username string) {
	id := uuid.NewString()
	ttl := domain.FifteenMinutesTime

	if err := s.storeEmailResetPassword(ctx, email, id, ttl); err != nil {
		pkg.Log().Errorw("[CACHE ERROR]", "from", "email_forgot_password", "error", err)
		return
	}

	data := domain.EventEmailData{
		Email:  email,
		Name:   username,
		Link:   s.url.ResetPasswordLink(id),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}

	payload := domain.EventPayload{
		EventID:   uuid.NewString(),
		EventType: domain.EmailResetPassword,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	if err := s.event.Publish(ctx, rabbitmq.EmailResetPasswordQueueConfig.RoutingKey, payload); err != nil {
		pkg.Log().Errorw("[EVENT QUEUE ERROR]", "from", "email_forgot_password", "error", err)
	}
}

func (s *AuthServiceImpl) storeEmailResetPassword(ctx context.Context, email, token string, ttl time.Duration) error {
	key := domain.GetEmailResetPasswordKey(token)
	return s.cache.Set(ctx, key, email, ttl)
}

func (s *AuthServiceImpl) getEmailResetPassword(ctx context.Context, token string) (string, error) {
	var email string
	key := domain.GetEmailResetPasswordKey(token)
	if err := s.cache.Get(ctx, key, &email); err != nil {
		return "", err
	}
	return email, nil
}
