package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/pkg"
)

type TokenService interface {
	GenerateTokens(ctx context.Context, userID int64) (*domain.TokenInfo, error)
	Refresh(ctx context.Context, raw string) (*domain.TokenInfo, error)
	Revoke(ctx context.Context, raw string) error
	Validate(access string) (*jwt.Token, error)
}

type TokenServiceImpl struct {
	repo domain.TokenRepository
	cfg  config.TokenConfig
}

func NewTokenService(repo domain.TokenRepository, cfg config.TokenConfig) *TokenServiceImpl {
	return &TokenServiceImpl{repo: repo, cfg: cfg}
}

func (s *TokenServiceImpl) generateAccessToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"aud": s.cfg.Aud,
		"iss": s.cfg.Iss,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(s.cfg.AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *TokenServiceImpl) generateRefreshToken(userID int64) (string, *domain.RefreshToken) {
	raw := uuid.NewString()
	h := sha256.Sum256([]byte(raw))
	hashed := hex.EncodeToString(h[:])

	return raw, &domain.RefreshToken{
		UserID:    userID,
		TokenHash: hashed,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (s *TokenServiceImpl) GenerateTokens(ctx context.Context, userID int64) (*domain.TokenInfo, error) {
	access, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	raw, refresh := s.generateRefreshToken(userID)
	if err := s.repo.Create(ctx, refresh); err != nil {
		return nil, err
	}

	return &domain.TokenInfo{
		AccessToken:  access,
		RefreshToken: hashToken(raw),
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *TokenServiceImpl) Refresh(ctx context.Context, raw string) (*domain.TokenInfo, error) {
	hash := hashToken(raw)

	dbToken, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		return nil, pkg.ErrUnauthorized
	}

	// Check expiration
	if dbToken.ExpiresAt.Before(time.Now()) {
		return nil, pkg.ErrUnauthorized
	}

	// Rotate token (revoke old)
	_ = s.repo.Revoke(ctx, dbToken.ID)

	// New token pair
	return s.GenerateTokens(ctx, dbToken.UserID)
}

func (s *TokenServiceImpl) Revoke(ctx context.Context, raw string) error {
	hash := hashToken(raw)
	dbToken, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		return nil
	}
	return s.repo.Revoke(ctx, dbToken.ID)
}

func (s *TokenServiceImpl) Validate(access string) (*jwt.Token, error) {
	return jwt.Parse(access, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(s.cfg.Aud),
		jwt.WithIssuer(s.cfg.Iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
