package verify

import (
	"context"
	"time"

	"github.com/google/uuid"

	"air-social/internal/cache"
	"air-social/internal/domain/common"
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
	ValidateResetPasswordToken(ctx context.Context, emailToken string) bool
}

type Deps struct {
	Cache cache.AtomicCache[string]
	Event common.EventPublisher
	Link  common.LinkProvider
}

type provider struct {
	cache cache.AtomicCache[string]
	event common.EventPublisher
	link  common.LinkProvider
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

	payload := common.EmailEventPayload{
		Email:  email,
		Name:   username,
		Link:   p.link.VerifyEmail(id),
		Expiry: pkg.FormatTTLHuman(verificationTTL),
	}

	return p.event.Publish(ctx, common.NewEvent(common.EventEmailVerify, payload))
}

func (p *provider) SendPasswordReset(ctx context.Context, email string, username string) error {
	id := uuid.NewString()

	if err := p.cache.Set(ctx, getResetKey(id), email, resetPasswordTTL); err != nil {
		return pkg.ErrInternal
	}

	payload := common.EmailEventPayload{
		Email:  email,
		Name:   username,
		Link:   p.link.ResetPassword(id),
		Expiry: pkg.FormatTTLHuman(resetPasswordTTL),
	}

	return p.event.Publish(ctx, common.NewEvent(common.EventEmailResetPassword, payload))
}

func (p *provider) VerifyVerification(ctx context.Context, token string) (string, error) {
	return p.cache.Get(ctx, getVerifyKey(token))
}

func (p *provider) VerifyPasswordReset(ctx context.Context, token string) (string, error) {
	return p.cache.Get(ctx, getResetKey(token))
}

func (p *provider) InvalidatePasswordReset(ctx context.Context, token string) error {
	return p.cache.Delete(ctx, getResetKey(token))
}

func (p *provider) ValidateResetPasswordToken(ctx context.Context, token string) bool {
	exists, err := p.cache.Exists(ctx, getResetKey(token))
	if err != nil {
		return false
	}
	return exists
}

func getVerifyKey(token string) string {
	return common.BuildCacheKey("auth", "email", "verify", token)
}

func getResetKey(token string) string {
	return common.BuildCacheKey("auth", "email", "reset", token)
}
