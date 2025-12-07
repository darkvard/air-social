package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/pkg"
)

type TokenService interface {
	CreateSession(ctx context.Context, userID int64, deviceID string) (*domain.TokenInfo, error)
	Refresh(ctx context.Context, accessToken, refreshToken string) (*domain.TokenInfo, error)
	RevokeSingle(ctx context.Context, refreshToken string) error
	RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	CleanupDatabase(ctx context.Context) error
	Validate(accessToken string) (*jwt.Token, error)
	Block(ctx context.Context, accessToken string) error
	IsBlocked(ctx context.Context, accessToken string) (bool, error)
}

type TokenServiceImpl struct {
	repo  domain.TokenRepository
	cfg   config.TokenConfig
	redis cache.CacheStorage
}

func NewTokenService(repo domain.TokenRepository, cfg config.TokenConfig, redis cache.CacheStorage) *TokenServiceImpl {
	return &TokenServiceImpl{repo: repo, cfg: cfg, redis: redis}
}

func (s *TokenServiceImpl) CreateSession(ctx context.Context, userID int64, deviceID string) (*domain.TokenInfo, error) {
	if err := s.RevokeDeviceSession(ctx, userID, deviceID); err != nil {
		pkg.Log().Warnw("failed to revoke existing session on login", "user_id", userID, "device_id", deviceID, "error", err)
	}
	return s.generateTokens(ctx, userID, deviceID)
}

func (s *TokenServiceImpl) generateTokens(ctx context.Context, userID int64, deviceID string) (*domain.TokenInfo, error) {
	access, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	raw, refresh := s.generateRefreshToken(userID, deviceID)
	if err := s.repo.Create(ctx, refresh); err != nil {
		return nil, err
	}

	return &domain.TokenInfo{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		TokenType:    pkg.AuthorizationType,
	}, nil
}

func (s *TokenServiceImpl) generateAccessToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		pkg.JWTClaimSubject:   fmt.Sprintf("%d", userID),
		pkg.JWTClaimID:        uuid.NewString(),
		pkg.JWTClaimAudience:  s.cfg.Aud,
		pkg.JWTClaimIssuer:    s.cfg.Iss,
		pkg.JWTClaimIssuedAt:  now.Unix(),
		pkg.JWTClaimNotBefore: now.Unix(),
		pkg.JWTClaimExpiresAt: now.Add(s.cfg.AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *TokenServiceImpl) generateRefreshToken(userID int64, deviceID string) (string, *domain.RefreshToken) {
	raw := uuid.NewString()
	h := sha256.Sum256([]byte(raw))
	hashed := hex.EncodeToString(h[:])

	return raw, &domain.RefreshToken{
		UserID:    userID,
		DeviceID:  deviceID,
		TokenHash: hashed,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenTTL),
		CreatedAt: time.Now(),
	}
}

func (s *TokenServiceImpl) hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (s *TokenServiceImpl) Refresh(ctx context.Context, accessToken, refreshToken string) (*domain.TokenInfo, error) {
	dbToken, err := s.verifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	newTokens, err := s.rotateSession(ctx, dbToken)
	if err != nil {
		return nil, err
	}

	if len(accessToken) > 0 {
		if err := s.Block(ctx, accessToken); err != nil {
			pkg.Log().Warnw("refresh: failed to block old access token", "error", err)
		}
	}

	return newTokens, nil
}

func (s *TokenServiceImpl) verifyRefreshToken(ctx context.Context, rawRefreshToken string) (*domain.RefreshToken, error) {
	dbToken, err := s.repo.GetByHash(ctx, s.hashToken(rawRefreshToken))
	if err != nil {
		pkg.Log().Infow("refresh: token not found or db error", "error", err)
		return nil, pkg.ErrUnauthorized
	}

	if dbToken.RevokedAt != nil {
		pkg.Log().Warnw("SECURITY ALERT: Reuse of revoked refresh token detected",
			"user_id", dbToken.UserID,
			"token_id", dbToken.ID,
			"device_id", dbToken.DeviceID,
		)
		_ = s.repo.UpdateRevokedByUser(ctx, dbToken.UserID)
		return nil, pkg.ErrUnauthorized
	}

	if dbToken.ExpiresAt.Before(time.Now()) {
		return nil, pkg.ErrUnauthorized
	}
	return dbToken, nil
}

func (s *TokenServiceImpl) rotateSession(ctx context.Context, oldToken *domain.RefreshToken) (*domain.TokenInfo, error) {
	if err := s.repo.UpdateRevoked(ctx, oldToken.ID); err != nil {
		pkg.Log().Errorw("CRITICAL: refresh: failed to revoke used token", "token_id", oldToken.ID, "error", err)
		return nil, err
	}

	return s.generateTokens(ctx, oldToken.UserID, oldToken.DeviceID)
}

func (s *TokenServiceImpl) RevokeSingle(ctx context.Context, refreshToken string) error {
	dbToken, err := s.repo.GetByHash(ctx, s.hashToken(refreshToken))
	if err != nil {
		return err
	}
	return s.repo.UpdateRevoked(ctx, dbToken.ID)
}

func (s *TokenServiceImpl) RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error {
	return s.repo.UpdateRevokedByDevice(ctx, userID, deviceID)
}

func (s *TokenServiceImpl) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	return s.repo.UpdateRevokedByUser(ctx, userID)
}

func (s *TokenServiceImpl) CleanupDatabase(ctx context.Context) error {
	threshold := time.Now().Add(-domain.AuditRetentionPeriod)
	return s.repo.DeleteExpiredAndRevoked(ctx, threshold, threshold)
}

func (s *TokenServiceImpl) Validate(accessToken string) (*jwt.Token, error) {
	return jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
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

func (s *TokenServiceImpl) Block(ctx context.Context, accessToken string) error {
	var jti string
	if err := pkg.ExtractClaimFromString(accessToken, pkg.JWTClaimID, &jti); err != nil {
		return err
	}
	if len(jti) == 0 {
		return errors.New("token invalid: jti claim is missing or empty")
	}
	return s.redis.Set(ctx, domain.BlockedAccessTokenKey(jti), true, s.cfg.AccessTokenTTL)
}

func (s *TokenServiceImpl) IsBlocked(ctx context.Context, accessToken string) (bool, error) {
	var jti string
	if err := pkg.ExtractClaimFromString(accessToken, pkg.JWTClaimID, &jti); err != nil {
		return false, err
	}
	if len(jti) == 0 {
		return false, errors.New("token invalid: jti claim is missing or empty")
	}
	return s.redis.IsExist(ctx, domain.BlockedAccessTokenKey(jti))
}
