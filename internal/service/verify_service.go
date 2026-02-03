package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain"
	"air-social/pkg"
)

type VerifyService interface {
	SendEmailVerification(ctx context.Context, email, username string) error
	VerifyEmailToken(ctx context.Context, token string) (string, error)

	SendPasswordReset(ctx context.Context, email, username string) error
	VerifyPasswordResetToken(ctx context.Context, token string) (string, error)
	InvalidatePasswordResetToken(ctx context.Context, token string) error
}

type VerifyServiceImpl struct {
	cache domain.CacheStorage
	event domain.EventPublisher
	url   domain.URLFactory
}

func NewVerifyService(cache domain.CacheStorage, event domain.EventPublisher, url domain.URLFactory) *VerifyServiceImpl {
	return &VerifyServiceImpl{
		cache: cache,
		event: event,
		url:   url,
	}
}

func (s *VerifyServiceImpl) SendEmailVerification(ctx context.Context, email, username string) error {
	id := uuid.NewString()
	ttl := 30 * time.Minute

	if err := s.cache.Set(ctx, domain.GetEmailVerificationKey(id), email, ttl); err != nil {
		return err
	}

	data := domain.EmailEvent{
		Email:  email,
		Name:   username,
		Link:   s.url.VerifyEmailLink(id),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}
	payload := domain.Event{
		EventID:   id,
		EventType: domain.EmailVerify,
		Timestamp: pkg.TimeNowUTC(),
		Data:      data,
	}

	return s.event.Publish(ctx, string(domain.EmailVerify), payload)
}

func (s *VerifyServiceImpl) VerifyEmailToken(ctx context.Context, token string) (string, error) {
	var email string
	key := domain.GetEmailVerificationKey(token)
	if err := s.cache.Get(ctx, key, &email); err != nil {
		return "", err
	}
	return email, nil
}

func (s *VerifyServiceImpl) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	key := domain.GetEmailResetPasswordKey(token)
	return s.cache.Delete(ctx, key)
}

func (s *VerifyServiceImpl) SendPasswordReset(ctx context.Context, email, username string) error {
	id := uuid.NewString()
	ttl := 15 * time.Minute

	if err := s.cache.Set(ctx, domain.GetEmailResetPasswordKey(id), email, ttl); err != nil {
		return err
	}

	data := domain.EmailEvent{
		Email:  email,
		Name:   username,
		Link:   s.url.ResetPasswordLink(id),
		Expiry: pkg.FormatTTLVerbose(ttl),
	}

	payload := domain.Event{
		EventID:   uuid.NewString(),
		EventType: domain.EmailResetPassword,
		Timestamp: pkg.TimeNowUTC(),
		Data:      data,
	}

	return s.event.Publish(ctx, string(domain.EmailResetPassword), payload)
}

func (s *VerifyServiceImpl) VerifyPasswordResetToken(ctx context.Context, token string) (string, error) {
	var email string
	key := domain.GetEmailResetPasswordKey(token)
	if err := s.cache.Get(ctx, key, &email); err != nil {
		return "", err
	}
	return email, nil
}
