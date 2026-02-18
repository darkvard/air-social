package verify

import (
	"context"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain/shared"
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
	Cache shared.Cache
	Event shared.EventPublisher
	Link  shared.AppLinkProvider
}

type provider struct {
	cache shared.Cache
	event shared.EventPublisher
	link  shared.AppLinkProvider
}

func NewVerifyProvider(d Deps) *provider {
	return &provider{
		cache: d.Cache,
		event: d.Event,
		link:  d.Link,
	}
}

func (p *provider) SendVerification(ctx context.Context, email string, username string) error {
	id := uuid.NewString()

	if err := p.cache.Set(ctx, getVerifyKey(id), email, verificationTTL); err != nil {
		return pkg.ErrInternal
	}

	payload := shared.EmailEventPayload{
		Email:  email,
		Name:   username,
		Link:   p.link.VerifyEmail(id),
		Expiry: pkg.FormatTTLHuman(verificationTTL),
	}
	event := shared.Event{
		ID:        id,
		Typ:       shared.EventVerify,
		Timestamp: pkg.TimeNowUTC(),
		Data:      payload,
	}

	return p.event.Publish(ctx, event)
}

func (p *provider) SendPasswordReset(ctx context.Context, email string, username string) error {
	id := uuid.NewString()

	if err := p.cache.Set(ctx, getResetKey(id), email, resetPasswordTTL); err != nil {
		return pkg.ErrInternal
	}

	payload := shared.EmailEventPayload{
		Email:  email,
		Name:   username,
		Link:   p.link.ResetPassword(id),
		Expiry: pkg.FormatTTLHuman(resetPasswordTTL),
	}
	event := shared.Event{
		ID:        id,
		Typ:       shared.EventResetPassword,
		Timestamp: pkg.TimeNowUTC(),
		Data:      payload,
	}

	return p.event.Publish(ctx, event)
}

func (p *provider) VerifyVerification(ctx context.Context, token string) (string, error) {
	var email string
	if err := p.cache.Get(ctx, getVerifyKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (p *provider) VerifyPasswordReset(ctx context.Context, token string) (string, error) {
	var email string
	if err := p.cache.Get(ctx, getResetKey(token), &email); err != nil {
		return "", err
	}
	return email, nil
}

func (p *provider) InvalidatePasswordReset(ctx context.Context, token string) error {
	return p.cache.Delete(ctx, getResetKey(token))
}

func getVerifyKey(token string) string {
	return shared.BuildCacheKey("worker", "email", "verify", token)
}

func getResetKey(token string) string {
	return shared.BuildCacheKey("worker", "email", "reset", token)
}
