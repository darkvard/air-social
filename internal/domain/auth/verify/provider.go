package verify

import (
	"context"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain"
	"air-social/pkg"
)

const (
	verificationTTL  = 30 * time.Minute
	resetPasswordTTL = 15 * time.Minute
)

type Provider interface {
	SendVerification(ctx context.Context, email, username string) error
	SendPasswordReset(ctx context.Context, email, username string) error

	VerifyVerification(ctx context.Context, token string) (string, error)
	VerifyPasswordReset(ctx context.Context, token string) (string, error)

	InvalidatePasswordReset(ctx context.Context, token string) error
}

type Deps struct {
	Cache domain.CacheStorage
	Event domain.EventPublisher
	URL   domain.URLFactory
}

type provider struct {
	cache domain.CacheStorage
	event domain.EventPublisher
	url   domain.URLFactory
}

func NewVerifyProvider(d Deps) *provider {
	return &provider{
		cache: d.Cache,
		event: d.Event,
		url:   d.URL,
	}
}

func (p *provider) SendVerification(ctx context.Context, email string, username string) error {
	id := uuid.NewString()

	if err := p.cache.Set(ctx, domain.GetEmailVerificationKey(id), email, verificationTTL); err != nil {
		return pkg.ErrInternal
	}

	return p.event.Publish(ctx, domain.RoutingKeyEmailVerify, domain.Event{
		EventID:   id,
		EventType: domain.EmailVerify,
		Timestamp: pkg.TimeNowUTC(),
		Data: domain.EmailEvent{
			Email:  email,
			Name:   username,
			Link:   p.url.VerifyEmailURL(id),
			Expiry: pkg.FormatTTLVerbose(verificationTTL),
		},
	})

}

func (p *provider) SendPasswordReset(ctx context.Context, email string, username string) error {
	id := uuid.NewString()

	if err := p.cache.Set(ctx, domain.GetEmailResetPasswordKey(id), email, resetPasswordTTL); err != nil {
		return pkg.ErrInternal
	}

	return p.event.Publish(ctx, domain.RoutingKeyEmailResetPassword, domain.Event{
		EventID:   id,
		EventType: domain.EmailResetPassword,
		Timestamp: pkg.TimeNowUTC(),
		Data: domain.EmailEvent{
			Email:  email,
			Name:   username,
			Link:   p.url.ResetPasswordURL(id),
			Expiry: pkg.FormatTTLVerbose(resetPasswordTTL),
		},
	})
}

func (p *provider) VerifyVerification(ctx context.Context, token string) (string, error) {
	var email string
	if err := p.cache.Get(ctx, domain.GetEmailVerificationKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (p *provider) VerifyPasswordReset(ctx context.Context, token string) (string, error) {
	var email string
	if err := p.cache.Get(ctx, domain.GetEmailResetPasswordKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (p *provider) InvalidatePasswordReset(ctx context.Context, token string) error {
	return p.cache.Delete(ctx, domain.GetEmailResetPasswordKey(token))
}
