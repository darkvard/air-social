package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/pkg"
)

type tokenProviderSuite struct {
	suite.Suite
	cfg config.TokenConfig
}

func TestTokenProviderSuite(t *testing.T) {
	suite.Run(t, new(tokenProviderSuite))
}

func (s *tokenProviderSuite) SetupSuite() {
	s.cfg = config.TokenConfig{
		Secret:          "secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Aud:             "test-aud",
		Iss:             "test-iss",
	}
}

func (s *tokenProviderSuite) TestGenerateAccessToken() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := token.NewProvider(s.cfg, mockCache)

	res, err := p.GenerateAccessToken(1, "device-1")

	s.NoError(err)
	s.NotEmpty(res.Token)
	s.WithinDuration(pkg.TimeNowUTC().Add(s.cfg.AccessTokenTTL), res.ExpiresAt, time.Second)
}

func (s *tokenProviderSuite) TestVerifyAccessToken() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := token.NewProvider(s.cfg, mockCache)

	// Generate valid token
	res, _ := p.GenerateAccessToken(1, "device-1")

	claims, access, err := p.VerifyAccessToken(res.Token)

	s.NoError(err)
	s.Equal(int64(1), claims.UserID)
	s.Equal("device-1", claims.DeviceID)
	s.Equal(res.Token, access.Token)
}

func (s *tokenProviderSuite) TestVerifyAccessToken_Invalid() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := token.NewProvider(s.cfg, mockCache)

	_, _, err := p.VerifyAccessToken("invalid.token.string")
	s.ErrorIs(err, pkg.ErrUnauthorized)
}

func (s *tokenProviderSuite) TestIsBlacklisted() {
	tokenStr := "access-token"

	type testDeps struct {
		cache *commonmocks.MockCache
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		want      bool
	}{
		{
			name: "blacklisted",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(true, nil).Once()
			},
			want: true,
		},
		{
			name: "not_blacklisted",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
			want: false,
		},
		{
			name: "cache_error",
			setupMock: func(deps testDeps) {
				deps.cache.EXPECT().
					IsExist(mock.Anything, mock.Anything).
					Return(false, assert.AnError).Once()
			},
			want: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := commonmocks.NewMockCache(s.T())
			p := token.NewProvider(s.cfg, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(testDeps{cache: mockCache})
			}

			got := p.IsBlacklisted(context.Background(), tokenStr)
			s.Equal(tc.want, got)
		})
	}
}

func (s *tokenProviderSuite) TestAddToBlacklist() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := token.NewProvider(s.cfg, mockCache)

	tokenStr := "token"
	expiresAt := pkg.TimeNowUTC().Add(1 * time.Hour)

	mockCache.EXPECT().
		Set(mock.Anything, mock.Anything, "revoked", mock.Anything).
		Return(nil).Once()

	p.AddToBlacklist(context.Background(), tokenStr, expiresAt)
}

func (s *tokenProviderSuite) TestVerifyRefreshToken() {
	mockCache := commonmocks.NewMockCache(s.T())
	p := token.NewProvider(s.cfg, mockCache)

	// Valid
	validToken := token.RefreshToken{
		ExpiresAt: pkg.TimeNowUTC().Add(time.Hour),
	}
	ok, err := p.VerifyRefreshToken(validToken)
	s.NoError(err)
	s.True(ok)

	// Revoked
	revokedTime := pkg.TimeNowUTC()
	revokedToken := token.RefreshToken{
		RevokedAt: &revokedTime,
	}
	ok, err = p.VerifyRefreshToken(revokedToken)
	s.ErrorIs(err, pkg.ErrTokenRevoked)
	s.False(ok)

	// Expired
	expiredToken := token.RefreshToken{ExpiresAt: pkg.TimeNowUTC().Add(-time.Hour)}
	ok, err = p.VerifyRefreshToken(expiredToken)
	s.ErrorIs(err, pkg.ErrUnauthorized)
	s.False(ok)
}
