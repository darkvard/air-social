package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/pkg"
)

type tokenServiceSuite struct {
	suite.Suite
	cfg config.TokenConfig
}

func TestTokenServiceSuite(t *testing.T) {
	suite.Run(t, new(tokenServiceSuite))
}

func (s *tokenServiceSuite) SetupSuite() {
	s.cfg = config.TokenConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Secret:          "secret",
		Aud:             "users",
		Iss:             "air-social",
	}
}

func (s *tokenServiceSuite) TestCreateSession() {
	var (
		userID   int64 = 1
		deviceID       = "device-123"
	)

	type args struct {
		userID   int64
		deviceID string
	}

	type want struct {
		tokenInfo domain.TokenInfo
		err       error
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(repo *mocks.TokenRepository)
		want      want
	}{
		{
			name: "revoke_device_error_ignored",
			args: args{userID: userID, deviceID: deviceID},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(assert.AnError).Once()
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Once()
			},
			want: want{
				tokenInfo: domain.TokenInfo{
					TokenType: pkg.AuthorizationType,
					ExpiresIn: int64(s.cfg.AccessTokenTTL.Seconds()),
				},
				err: nil,
			},
		},
		{
			name: "create_token_error",
			args: args{userID: userID, deviceID: deviceID},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(nil).Once()
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			want: want{
				err: pkg.ErrInternal,
			},
		},
		{
			name: "success",
			args: args{userID: userID, deviceID: deviceID},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(nil).Once()
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t domain.RefreshToken) bool {
					return t.UserID == userID && t.DeviceID == deviceID
				})).Return(nil).Once()
			},
			want: want{
				tokenInfo: domain.TokenInfo{
					TokenType: pkg.AuthorizationType,
					ExpiresIn: int64(s.cfg.AccessTokenTTL.Seconds()),
				},
				err: nil,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			got, err := svc.CreateSession(context.Background(), tc.args.userID, tc.args.deviceID)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.NotEmpty(got.AccessToken)
				s.NotEmpty(got.RefreshToken)
				s.Equal(tc.want.tokenInfo.TokenType, got.TokenType)
				s.Equal(tc.want.tokenInfo.ExpiresIn, got.ExpiresIn)
			}
		})
	}
}

func (s *tokenServiceSuite) TestRefresh() {
	svc := NewTokenService(nil, s.cfg)
	rawToken := "raw-refresh-token"
	hashedToken := svc.hashToken(rawToken)

	dbToken := domain.RefreshToken{
		ID:        1,
		UserID:    1,
		DeviceID:  "device-1",
		TokenHash: hashedToken,
		ExpiresAt: pkg.TimeNowUTC().Add(1 * time.Hour),
	}

	type args struct {
		refreshToken string
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(repo *mocks.TokenRepository)
		wantErr   error
	}{
		{
			name: "token_not_found",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(domain.RefreshToken{}, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "token_revoked",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				revokedToken := dbToken
				now := pkg.TimeNowUTC()
				revokedToken.RevokedAt = &now
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(revokedToken, nil).Once()
				repo.EXPECT().UpdateRevokedByUser(mock.Anything, revokedToken.UserID).Return(nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "token_expired",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				expiredToken := dbToken
				expiredToken.ExpiresAt = pkg.TimeNowUTC().Add(-1 * time.Hour)
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(expiredToken, nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "rotate_update_revoked_error",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(dbToken, nil).Once()
				repo.EXPECT().UpdateRevoked(mock.Anything, dbToken.ID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "rotate_create_error",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(dbToken, nil).Once()
				repo.EXPECT().UpdateRevoked(mock.Anything, dbToken.ID).Return(nil).Once()
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{refreshToken: rawToken},
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(dbToken, nil).Once()
				repo.EXPECT().UpdateRevoked(mock.Anything, dbToken.ID).Return(nil).Once()
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(t domain.RefreshToken) bool {
					return t.UserID == dbToken.UserID && t.DeviceID == dbToken.DeviceID
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)

			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			got, err := svc.Refresh(context.Background(), tc.args.refreshToken)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.NotEmpty(got.AccessToken)
				s.NotEmpty(got.RefreshToken)
			}
		})
	}
}

func (s *tokenServiceSuite) TestRevokeSingle() {
	svc := NewTokenService(nil, s.cfg)
	rawToken := "raw-token"
	hashedToken := svc.hashToken(rawToken)
	dbToken := domain.RefreshToken{ID: 1}

	tests := []struct {
		name      string
		token     string
		setupMock func(repo *mocks.TokenRepository)
		wantErr   error
	}{
		{
			name:  "get_error",
			token: rawToken,
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(domain.RefreshToken{}, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:  "update_error",
			token: rawToken,
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(dbToken, nil).Once()
				repo.EXPECT().UpdateRevoked(mock.Anything, dbToken.ID).Return(assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:  "success",
			token: rawToken,
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(dbToken, nil).Once()
				repo.EXPECT().UpdateRevoked(mock.Anything, dbToken.ID).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}
			err := svc.RevokeSingle(context.Background(), tc.token)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *tokenServiceSuite) TestRevokeDeviceSession() {
	var userID int64 = 1
	deviceID := "device-1"

	tests := []struct {
		name      string
		setupMock func(repo *mocks.TokenRepository)
		wantErr   error
	}{
		{
			name: "error",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}
			err := svc.RevokeDeviceSession(context.Background(), userID, deviceID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *tokenServiceSuite) TestRevokeAllUserSessions() {
	var userID int64 = 1

	tests := []struct {
		name      string
		setupMock func(repo *mocks.TokenRepository)
		wantErr   error
	}{
		{
			name: "error",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByUser(mock.Anything, userID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().UpdateRevokedByUser(mock.Anything, userID).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}
			err := svc.RevokeAllUserSessions(context.Background(), userID)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *tokenServiceSuite) TestCleanupDatabase() {
	tests := []struct {
		name      string
		setupMock func(repo *mocks.TokenRepository)
		wantErr   error
	}{
		{
			name: "error",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().DeleteExpiredAndRevoked(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			setupMock: func(repo *mocks.TokenRepository) {
				repo.EXPECT().DeleteExpiredAndRevoked(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewTokenRepository(s.T())
			svc := NewTokenService(mockRepo, s.cfg)
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}
			err := svc.CleanupDatabase(context.Background())
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *tokenServiceSuite) TestValidate() {
	svc := NewTokenService(nil, s.cfg)
	validToken, _ := svc.generateAccessToken(1, "device-1")

	tests := []struct {
		name        string
		tokenString string
		cfg         config.TokenConfig
		wantErr     error
		wantToken   bool
	}{
		{
			name:        "valid_token",
			tokenString: validToken,
			cfg:         s.cfg,
			wantErr:     nil,
			wantToken:   true,
		},
		{
			name:        "invalid_signature",
			tokenString: validToken,
			cfg: func() config.TokenConfig {
				c := s.cfg
				c.Secret = "wrong-secret"
				return c
			}(),
			wantErr:   jwt.ErrTokenSignatureInvalid,
			wantToken: true,
		},
		{
			name: "expired_token",
			tokenString: func() string {
				expiredCfg := s.cfg
				expiredCfg.AccessTokenTTL = -1 * time.Hour
				expiredSvc := NewTokenService(nil, expiredCfg)
				t, _ := expiredSvc.generateAccessToken(1, "device-1")
				return t
			}(),
			cfg:       s.cfg,
			wantErr:   jwt.ErrTokenExpired,
			wantToken: true,
		},
		{
			name:        "malformed_token",
			tokenString: "invalid-token-string",
			cfg:         s.cfg,
			wantErr:     jwt.ErrTokenMalformed,
			wantToken:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			svc := NewTokenService(nil, tc.cfg)
			token, err := svc.Validate(tc.tokenString)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				if tc.wantToken {
					s.NotNil(token)
					s.False(token.Valid)
				} else {
					s.Nil(token)
				}
			} else {
				s.NoError(err)
				s.NotNil(token)
				s.True(token.Valid)
			}
		})
	}
}
