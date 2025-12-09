package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/pkg"
)

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *MockTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) UpdateRevoked(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockTokenRepository) UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error {
	return m.Called(ctx, userID, deviceID).Error(0)
}

func (m *MockTokenRepository) UpdateRevokedByUser(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockTokenRepository) DeleteExpiredAndRevoked(ctx context.Context, expiredThreshold, revokedThreshold time.Time) error {
	return m.Called(ctx, expiredThreshold, revokedThreshold).Error(0)
}

type MockCacheStorage struct {
	mock.Mock
}

func (m *MockCacheStorage) Get(ctx context.Context, key string, dst any) error {
	return m.Called(ctx, key, dst).Error(0)
}

func (m *MockCacheStorage) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	return m.Called(ctx, key, val, ttl).Error(0)
}

func (m *MockCacheStorage) Delete(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockCacheStorage) IsExist(ctx context.Context, key string) (bool, error) {
	return m.Called(ctx, key).Get(0).(bool), m.Called(ctx, key).Error(1)
}

func TestTokenService_CreateSession(t *testing.T) {
	mockRepo := new(MockTokenRepository)
	cfg := config.TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Secret:          "secret",
		Aud:             "users",
		Iss:             "air-social",
	}
	service := NewTokenService(mockRepo, cfg)

	var userID int64 = 1
	deviceID := "device-123"

	tests := []struct {
		name          string
		setupMocks    func()
		checkResult   func(t *testing.T, tokenInfo *domain.TokenInfo)
		expectedError error
	}{
		{
			name: "success",
			setupMocks: func() {
				mockRepo.On("UpdateRevokedByDevice", mock.Anything, userID, deviceID).Return(nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			checkResult: func(t *testing.T, tokenInfo *domain.TokenInfo) {
				assert.NotNil(t, tokenInfo)
				assert.NotEmpty(t, tokenInfo.AccessToken)
				assert.NotEmpty(t, tokenInfo.RefreshToken)
				assert.Equal(t, "Bearer", tokenInfo.TokenType)
				assert.Equal(t, int64(cfg.AccessTokenTTL.Seconds()), tokenInfo.ExpiresIn)
			},
			expectedError: nil,
		},
		{
			name: "create token in repo fails",
			setupMocks: func() {
				mockRepo.On("UpdateRevokedByDevice", mock.Anything, userID, deviceID).Return(nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			checkResult: func(t *testing.T, tokenInfo *domain.TokenInfo) {
				assert.Nil(t, tokenInfo)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.Calls = nil
			mockRepo.ExpectedCalls = nil
			tc.setupMocks()

			tokenInfo, err := service.CreateSession(context.Background(), userID, deviceID)
			assert.ErrorIs(t, err, tc.expectedError)
			tc.checkResult(t, tokenInfo)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_Refresh(t *testing.T) {
	mockRepo := new(MockTokenRepository)
	cfg := config.TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Secret:          "secret",
	}
	service := NewTokenService(mockRepo, cfg)

	rawToken := "raw-refresh-token"
	hashedToken := service.hashToken(rawToken)
	dbToken := &domain.RefreshToken{
		ID:        1,
		UserID:    1,
		DeviceID:  "device-123",
		TokenHash: hashedToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tests := []struct {
		name          string
		setupMocks    func()
		expectedError error
	}{

		{
			name: "token not found",
			setupMocks: func() {
				mockRepo.On("GetByHash", mock.Anything, hashedToken).Return(nil, pkg.ErrNotFound)
			},
			expectedError: pkg.ErrNotFound,
		},
		{
			name: "token revoked",
			setupMocks: func() {
				revokedToken := *dbToken
				revokedToken.RevokedAt = &time.Time{}
				mockRepo.On("GetByHash", mock.Anything, hashedToken).Return(&revokedToken, nil)
				mockRepo.On("UpdateRevokedByUser", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: pkg.ErrTokenRevoked,
		},
		{
			name: "token expired",
			setupMocks: func() {
				expiredToken := *dbToken
				expiredToken.ExpiresAt = time.Now().Add(-1 * time.Hour)
				mockRepo.On("GetByHash", mock.Anything, hashedToken).Return(&expiredToken, nil)
			},
			expectedError: pkg.ErrTokenExpired,
		},
		{
			name: "revoke failed",
			setupMocks: func() {
				mockRepo.On("GetByHash", mock.Anything, hashedToken).Return(dbToken, nil)
				mockRepo.On("UpdateRevoked", mock.Anything, dbToken.ID).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name: "success",
			setupMocks: func() {
				mockRepo.On("GetByHash", mock.Anything, hashedToken).Return(dbToken, nil)
				mockRepo.On("UpdateRevoked", mock.Anything, dbToken.ID).Return(nil)
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.Calls = nil
			mockRepo.ExpectedCalls = nil
			tc.setupMocks()

			tokenInfo, err := service.Refresh(context.Background(), rawToken)

			assert.ErrorIs(t, err, tc.expectedError)
			if tc.expectedError == nil {
				assert.NotNil(t, tokenInfo)
			} else {
				assert.Nil(t, tokenInfo)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTokenService_Validate(t *testing.T) {
	cfg := config.TokenConfig{
		AccessTokenTTL: 15 * time.Minute,
		Secret:         "my-super-secret-key",
		Aud:            "users",
		Iss:            "air-social",
	}
	service := NewTokenService(nil, cfg)

	validToken, _ := service.generateAccessToken(1, "test-device")

	tests := []struct {
		name          string
		tokenString   string
		service       *TokenServiceImpl
		expectedError error
	}{
		{
			name:          "valid token",
			tokenString:   validToken,
			service:       service,
			expectedError: nil,
		},
		{
			name:          "invalid signature",
			tokenString:   validToken,
			service:       NewTokenService(nil, config.TokenConfig{Secret: "wrong-secret"}),
			expectedError: jwt.ErrTokenSignatureInvalid,
		},
		{
			name: "expired token",
			tokenString: func() string {
				expiredCfg := cfg
				expiredCfg.AccessTokenTTL = -1 * time.Minute
				expiredService := NewTokenService(nil, expiredCfg)
				token, _ := expiredService.generateAccessToken(1, "test-device")
				return token
			}(),
			service:       service,
			expectedError: jwt.ErrTokenExpired,
		},
		{
			name:          "invalid signing method",
			tokenString:   jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"sub": "1"}).Raw,
			service:       service,
			expectedError: jwt.ErrTokenMalformed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := tc.service.Validate(tc.tokenString)
			assert.ErrorIs(t, err, tc.expectedError)
			if tc.expectedError == nil {
				assert.True(t, token.Valid)
			}
		})
	}
}

func TestTokenService_Revoke(t *testing.T) {
	ctx := context.Background()

	t.Run("RevokeSingle", func(t *testing.T) {
		mockRepo := new(MockTokenRepository)
		service := NewTokenService(mockRepo, config.TokenConfig{})

		mockRepo.On("GetByHash", ctx, mock.Anything).Return(&domain.RefreshToken{ID: 1}, nil)
		mockRepo.On("UpdateRevoked", ctx, int64(1)).Return(nil)

		err := service.RevokeSingle(ctx, "some-token")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RevokeDeviceSession", func(t *testing.T) {
		mockRepo := new(MockTokenRepository)
		service := NewTokenService(mockRepo, config.TokenConfig{})

		mockRepo.On("UpdateRevokedByDevice", ctx, int64(1), "device-1").Return(nil)

		err := service.RevokeDeviceSession(ctx, 1, "device-1")
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RevokeAllUserSessions", func(t *testing.T) {
		mockRepo := new(MockTokenRepository)
		service := NewTokenService(mockRepo, config.TokenConfig{})

		mockRepo.On("UpdateRevokedByUser", ctx, int64(1)).Return(nil)

		err := service.RevokeAllUserSessions(ctx, 1)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTokenService_CleanupDatabase(t *testing.T) {
	mockRepo := new(MockTokenRepository)
	service := NewTokenService(mockRepo, config.TokenConfig{})
	ctx := context.Background()

	mockRepo.On("DeleteExpiredAndRevoked", ctx, mock.Anything, mock.Anything).Return(nil)

	err := service.CleanupDatabase(ctx)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
