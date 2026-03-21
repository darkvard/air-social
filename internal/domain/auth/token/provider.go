package token

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain/common"
	"air-social/pkg"
)

type Provider interface {
	HashToken(raw string) string

	GenerateAccessToken(userID int64, deviceID string) (AccessTokenResult, error)
	GenerateRefreshToken() RefreshTokenResult
	VerifyAccessToken(token string) (TokenClaims, AccessTokenResult, error)
	VerifyRefreshToken(token RefreshToken) (bool, error)

	IsBlacklisted(ctx context.Context, accessToken string) bool
	AddToBlacklist(ctx context.Context, accessToken string, expiresAt time.Time)
}

type provider struct {
	cfg   config.TokenConfig
	cache cache.AtomicCache[string]
}

func NewProvider(cfg config.TokenConfig, cache cache.AtomicCache[string]) *provider {
	return &provider{
		cfg:   cfg,
		cache: cache,
	}
}

func (p *provider) HashToken(raw string) string {
	src := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(src[:])
}

func (p *provider) GenerateAccessToken(userID int64, deviceID string) (AccessTokenResult, error) {
	var empty AccessTokenResult

	now := pkg.TimeNowUTC()
	expiresAt := now.Add(p.cfg.AccessTokenTTL)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		pkg.JWTClaimSubject:   fmt.Sprintf("%d", userID),
		pkg.JWTClaimDevice:    deviceID,
		pkg.JWTClaimAudience:  p.cfg.Aud,
		pkg.JWTClaimIssuer:    p.cfg.Iss,
		pkg.JWTClaimIssuedAt:  now.Unix(),
		pkg.JWTClaimNotBefore: now.Unix(),
		pkg.JWTClaimExpiresAt: expiresAt.Unix(),
	}).SignedString([]byte(p.cfg.Secret))

	if err != nil {
		return empty, pkg.ErrInternal
	}

	return AccessTokenResult{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (p *provider) GenerateRefreshToken() RefreshTokenResult {
	raw := uuid.NewString()
	hashed := p.HashToken(raw)
	now := pkg.TimeNowUTC()
	expiresAt := now.Add(p.cfg.RefreshTokenTTL)

	return RefreshTokenResult{
		Raw:       raw,
		Hashed:    hashed,
		ExpiresAt: expiresAt,
	}
}

func (p *provider) VerifyAccessToken(token string) (TokenClaims, AccessTokenResult, error) {
	var emptyClaims TokenClaims
	var emptyAccess AccessTokenResult

	res, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(p.cfg.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(p.cfg.Aud),
		jwt.WithIssuer(p.cfg.Iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil || !res.Valid {
		return emptyClaims, emptyAccess, pkg.ErrUnauthorized
	}

	claims, ok := res.Claims.(jwt.MapClaims)
	if !ok {
		return emptyClaims, emptyAccess, pkg.ErrUnauthorized
	}

	tokenClaims := TokenClaims{
		UserID:   pkg.GetInt64Claims(claims, pkg.JWTClaimSubject),
		DeviceID: pkg.GetStringClaims(claims, pkg.JWTClaimDevice),
	}

	accessToken := AccessTokenResult{
		Token:     token,
		ExpiresAt: pkg.GetTimeClaims(claims, pkg.JWTClaimExpiresAt),
	}

	return tokenClaims, accessToken, nil

}

func (p *provider) VerifyRefreshToken(token RefreshToken) (bool, error) {
	if token.RevokedAt != nil {
		return false, pkg.ErrTokenRevoked
	}

	if token.ExpiresAt.Before(pkg.TimeNowUTC()) {
		return false, pkg.ErrUnauthorized
	}

	return true, nil
}

func (p *provider) IsBlacklisted(ctx context.Context, accessToken string) bool {
	exists, err := p.cache.Exists(ctx, getBlacklistTokenKey(accessToken))
	if err != nil {
		return false
	}
	return exists
}

func (p *provider) AddToBlacklist(ctx context.Context, accessToken string, expiresAt time.Time) {
	ttl := time.Until(expiresAt)
	if ttl > 0 {
		key := getBlacklistTokenKey(accessToken)
		_ = p.cache.Set(ctx, key, "revoked", ttl)
	}
}

func getBlacklistTokenKey(token string) string {
	return common.BuildCacheKey("auth", "blacklist", token)
}
