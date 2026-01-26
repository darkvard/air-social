package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/pkg"
)

type TokenService interface {
	CreateSession(ctx context.Context, userID int64, deviceID string) (domain.TokenInfo, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenInfo, error)
	RevokeSingle(ctx context.Context, refreshToken string) error
	RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	CleanupDatabase(ctx context.Context) error
	Validate(accessToken string) (*jwt.Token, error)
}

type TokenServiceImpl struct {
	tokenRepo domain.TokenRepository
	tokenCfg  config.TokenConfig
}

func NewTokenService(repo domain.TokenRepository, cfg config.TokenConfig) *TokenServiceImpl {
	return &TokenServiceImpl{tokenRepo: repo, tokenCfg: cfg}
}

func (s *TokenServiceImpl) CreateSession(ctx context.Context, userID int64, deviceID string) (domain.TokenInfo, error) {
	_ = s.RevokeDeviceSession(ctx, userID, deviceID)

	var empty domain.TokenInfo
	res, err := s.generateTokens(ctx, userID, deviceID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return res, nil
}

func (s *TokenServiceImpl) Refresh(ctx context.Context, refreshToken string) (domain.TokenInfo, error) {
	var empty domain.TokenInfo

	dbToken, err := s.verifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return empty, err
	}

	newTokens, err := s.rotateSession(ctx, dbToken)
	if err != nil {
		return empty, err
	}

	return newTokens, nil
}

func (s *TokenServiceImpl) RevokeSingle(ctx context.Context, refreshToken string) error {
	dbToken, err := s.tokenRepo.GetByHash(ctx, s.hashToken(refreshToken))
	if err != nil {
		return err
	}
	return s.tokenRepo.UpdateRevoked(ctx, dbToken.ID)
}

func (s *TokenServiceImpl) RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error {
	if err := s.tokenRepo.UpdateRevokedByDevice(ctx, userID, deviceID); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (s *TokenServiceImpl) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	if err := s.tokenRepo.UpdateRevokedByUser(ctx, userID); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (s *TokenServiceImpl) CleanupDatabase(ctx context.Context) error {
	threshold := pkg.TimeNowUTC().Add(-domain.AuditRetentionPeriod)
	if err := s.tokenRepo.DeleteExpiredAndRevoked(ctx, threshold, threshold); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (s *TokenServiceImpl) Validate(accessToken string) (*jwt.Token, error) {
	return jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(s.tokenCfg.Secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(s.tokenCfg.Aud),
		jwt.WithIssuer(s.tokenCfg.Iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}

// Internal helpers
func (s *TokenServiceImpl) verifyRefreshToken(ctx context.Context, rawRefreshToken string) (domain.RefreshToken, error) {
	var empty domain.RefreshToken
	dbToken, err := s.tokenRepo.GetByHash(ctx, s.hashToken(rawRefreshToken))
	if err != nil {
		return empty, pkg.ErrUnauthorized
	}

	if dbToken.RevokedAt != nil {
		_ = s.tokenRepo.UpdateRevokedByUser(ctx, dbToken.UserID)
		return empty, pkg.ErrUnauthorized
	}

	if dbToken.ExpiresAt.Before(pkg.TimeNowUTC()) {
		return empty, pkg.ErrUnauthorized
	}
	return dbToken, nil
}

func (s *TokenServiceImpl) rotateSession(ctx context.Context, oldToken domain.RefreshToken) (domain.TokenInfo, error) {
	var empty domain.TokenInfo
	if err := s.tokenRepo.UpdateRevoked(ctx, oldToken.ID); err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return s.generateTokens(ctx, oldToken.UserID, oldToken.DeviceID)
}

func (s *TokenServiceImpl) generateTokens(ctx context.Context, userID int64, deviceID string) (domain.TokenInfo, error) {
	var empty domain.TokenInfo

	access, err := s.generateAccessToken(userID, deviceID)
	if err != nil {
		return empty, pkg.OrInternalError(err)
	}

	raw, refresh := s.generateRefreshToken(userID, deviceID)

	if err := s.tokenRepo.Create(ctx, refresh); err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return domain.TokenInfo{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int64(s.tokenCfg.AccessTokenTTL.Seconds()),
		TokenType:    pkg.AuthorizationType,
	}, nil
}

func (s *TokenServiceImpl) generateAccessToken(userID int64, deviceID string) (string, error) {
	now := pkg.TimeNowUTC()
	claims := jwt.MapClaims{
		pkg.JWTClaimSubject:   fmt.Sprintf("%d", userID),
		pkg.JWTClaimDevice:    deviceID,
		pkg.JWTClaimAudience:  s.tokenCfg.Aud,
		pkg.JWTClaimIssuer:    s.tokenCfg.Iss,
		pkg.JWTClaimIssuedAt:  now.Unix(),
		pkg.JWTClaimNotBefore: now.Unix(),
		pkg.JWTClaimExpiresAt: now.Add(s.tokenCfg.AccessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.tokenCfg.Secret))
}

func (s *TokenServiceImpl) generateRefreshToken(userID int64, deviceID string) (string, domain.RefreshToken) {
	raw := uuid.NewString()
	hashed := s.hashToken(raw)
	now := pkg.TimeNowUTC()
	expiresAt := now.Add(s.tokenCfg.RefreshTokenTTL)

	return raw, domain.RefreshToken{
		UserID:    userID,
		DeviceID:  deviceID,
		TokenHash: hashed,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
}

func (s *TokenServiceImpl) hashToken(raw string) string {
	src := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(src[:])
}
