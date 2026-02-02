package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/pkg"
)

type authServiceSuite struct {
	suite.Suite
}

func TestAuthServiceSuite(t *testing.T) {
	suite.Run(t, new(authServiceSuite))
}

func (s *authServiceSuite) TestRegister() {
	input := domain.RegisterParams{
		Email:    "test@example.com",
		Username: "tester",
		Password: "password123",
	}

	userResp := &domain.User{
		ID:       1,
		Email:    input.Email,
		Username: input.Username,
	}

	type args struct {
		input domain.RegisterParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(u *mocks.UserService, t *mocks.TokenService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage)
		want      *domain.User
		wantErr   error
	}{
		{
			name: "create_user_error",
			args: args{input: input},
			setupMock: func(u *mocks.UserService, t *mocks.TokenService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage) {
				u.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil, pkg.ErrAlreadyExists).Once()
			},
			want:    nil,
			wantErr: pkg.ErrAlreadyExists,
		},
		{
			name: "success",
			args: args{input: input},
			setupMock: func(u *mocks.UserService, t *mocks.TokenService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage) {
				u.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(p domain.CreateUserParams) bool {
					return p.Email == input.Email && p.Username == input.Username && p.PasswordHashed != ""
				})).Return(userResp, nil).Once()

				// sendEmailVerification flow
				c.EXPECT().Set(mock.Anything, mock.Anything, input.Email, 30*time.Minute).Return(nil).Once()
				url.EXPECT().VerifyEmailLink(mock.Anything).Return("http://verify.link").Once()
				e.EXPECT().Publish(mock.Anything, mock.Anything, mock.MatchedBy(func(p domain.Event) bool {
					data, ok := p.Data.(domain.EmailEvent)
					return ok && data.Name == userResp.Username
				})).Return(nil).Once()
			},
			want:    userResp,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUser := mocks.NewUserService(s.T())
			mockToken := mocks.NewTokenService(s.T())
			mockURL := mocks.NewURLFactory(s.T())
			mockEvent := mocks.NewEventPublisher(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			svc := NewAuthService(mockUser, mockToken, mockURL, mockEvent, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockUser, mockToken, mockURL, mockEvent, mockCache)
			}

			got, err := svc.Register(context.Background(), tc.args.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *authServiceSuite) TestLogin() {
	password := "password123"
	hashedPwd, _ := hashPassword(password)

	input := domain.LoginParams{
		Email:    "test@example.com",
		Password: password,
		DeviceID: "device-1",
	}

	user := &domain.User{
		ID:           1,
		Email:        input.Email,
		Username:     "tester",
		PasswordHash: hashedPwd,
	}

	tokenInfo := domain.TokenInfo{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	tests := []struct {
		name      string
		input     domain.LoginParams
		setupMock func(u *mocks.UserService, t *mocks.TokenService)
		wantUser  *domain.User
		wantToken *domain.TokenInfo
		wantErr   error
	}{
		{
			name:  "user_not_found",
			input: input,
			setupMock: func(u *mocks.UserService, t *mocks.TokenService) {
				u.EXPECT().GetByEmail(mock.Anything, input.Email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name:  "invalid_password",
			input: input,
			setupMock: func(u *mocks.UserService, t *mocks.TokenService) {
				otherHash, _ := hashPassword("other")
				u.EXPECT().GetByEmail(mock.Anything, input.Email).Return(&domain.User{PasswordHash: otherHash}, nil).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name:  "token_creation_error",
			input: input,
			setupMock: func(u *mocks.UserService, t *mocks.TokenService) {
				u.EXPECT().GetByEmail(mock.Anything, input.Email).Return(user, nil).Once()
				t.EXPECT().CreateSession(mock.Anything, user.ID, input.DeviceID).Return(domain.TokenInfo{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:  "success",
			input: input,
			setupMock: func(u *mocks.UserService, t *mocks.TokenService) {
				u.EXPECT().GetByEmail(mock.Anything, input.Email).Return(user, nil).Once()
				t.EXPECT().CreateSession(mock.Anything, user.ID, input.DeviceID).Return(tokenInfo, nil).Once()
			},
			wantUser:  user,
			wantToken: &tokenInfo,
			wantErr:   nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUser := mocks.NewUserService(s.T())
			mockToken := mocks.NewTokenService(s.T())
			svc := NewAuthService(mockUser, mockToken, nil, nil, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockUser, mockToken)
			}

			gotUser, gotToken, err := svc.Login(context.Background(), tc.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.wantUser, gotUser)
				s.Equal(tc.wantToken, gotToken)
			}
		})
	}
}

func (s *authServiceSuite) TestLogout() {
	var userID int64 = 1
	deviceID := "device-1"
	tokenMeta := domain.TokenMeta{
		AccessToken: "access-token",
		ExpiresAt:   pkg.TimeNowUTC().Add(15 * time.Minute).Unix(),
	}

	tests := []struct {
		name      string
		input     domain.LogoutParams
		setupMock func(t *mocks.TokenService, c *mocks.CacheStorage)
		wantErr   error
	}{
		{
			name: "token_blocked",
			input: domain.LogoutParams{
				UserID:       userID,
				DeviceID:     deviceID,
				IsAllDevices: false,
				Token:        tokenMeta,
			},
			setupMock: func(t *mocks.TokenService, c *mocks.CacheStorage) {
				c.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(true, nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "logout_all_devices_success",
			input: domain.LogoutParams{
				UserID:       userID,
				IsAllDevices: true,
				Token:        tokenMeta,
			},
			setupMock: func(t *mocks.TokenService, c *mocks.CacheStorage) {
				c.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				t.EXPECT().RevokeAllUserSessions(mock.Anything, userID).Return(nil).Once()
				c.EXPECT().Set(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken), "revoked", mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "logout_single_device_success",
			input: domain.LogoutParams{
				UserID:       userID,
				DeviceID:     deviceID,
				IsAllDevices: false,
				Token:        tokenMeta,
			},
			setupMock: func(t *mocks.TokenService, c *mocks.CacheStorage) {
				c.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				t.EXPECT().RevokeDeviceSession(mock.Anything, userID, deviceID).Return(nil).Once()
				c.EXPECT().Set(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken), "revoked", mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "revoke_error",
			input: domain.LogoutParams{
				UserID:       userID,
				IsAllDevices: true,
				Token:        tokenMeta,
			},
			setupMock: func(t *mocks.TokenService, c *mocks.CacheStorage) {
				c.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				t.EXPECT().RevokeAllUserSessions(mock.Anything, userID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockToken := mocks.NewTokenService(s.T())
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewAuthService(nil, mockToken, nil, nil, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockToken, mockCache)
			}

			err := svc.Logout(context.Background(), tc.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authServiceSuite) TestForgotPassword() {
	email := "test@example.com"
	user := &domain.User{Email: email, Username: "tester"}

	tests := []struct {
		name      string
		email     string
		setupMock func(u *mocks.UserService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage)
		wantErr   error
	}{
		{
			name:  "user_not_found",
			email: email,
			setupMock: func(u *mocks.UserService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "success",
			email: email,
			setupMock: func(u *mocks.UserService, url *mocks.URLFactory, e *mocks.EventPublisher, c *mocks.CacheStorage) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(user, nil).Once()
				c.EXPECT().Set(mock.Anything, mock.Anything, email, 15*time.Minute).Return(nil).Once()
				url.EXPECT().ResetPasswordLink(mock.Anything).Return("http://reset.link").Once()
				e.EXPECT().Publish(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUser := mocks.NewUserService(s.T())
			mockURL := mocks.NewURLFactory(s.T())
			mockEvent := mocks.NewEventPublisher(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			svc := NewAuthService(mockUser, nil, mockURL, mockEvent, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockUser, mockURL, mockEvent, mockCache)
			}

			err := svc.ForgotPassword(context.Background(), tc.email)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authServiceSuite) TestResetPassword() {
	token := "reset-token"
	email := "test@example.com"
	input := domain.ResetPasswordParams{
		EmailToken: token,
		Password:   "newpassword",
	}

	tests := []struct {
		name      string
		input     domain.ResetPasswordParams
		setupMock func(u *mocks.UserService, c *mocks.CacheStorage)
		wantErr   error
	}{
		{
			name:  "token_invalid",
			input: input,
			setupMock: func(u *mocks.UserService, c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "success",
			input: input,
			setupMock: func(u *mocks.UserService, c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*string) = email
					}).Return(nil).Once()

				u.EXPECT().UpdatePassword(mock.Anything, email, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUser := mocks.NewUserService(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			svc := NewAuthService(mockUser, nil, nil, nil, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockUser, mockCache)
			}

			err := svc.ResetPassword(context.Background(), tc.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authServiceSuite) TestVerifyEmail() {
	token := "verify-token"
	email := "test@example.com"

	tests := []struct {
		name      string
		token     string
		setupMock func(u *mocks.UserService, c *mocks.CacheStorage)
		wantErr   error
	}{
		{
			name:  "token_invalid",
			token: token,
			setupMock: func(u *mocks.UserService, c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:  "success",
			token: token,
			setupMock: func(u *mocks.UserService, c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*string) = email
					}).Return(nil).Once()

				u.EXPECT().VerifyEmail(mock.Anything, email).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockUser := mocks.NewUserService(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			svc := NewAuthService(mockUser, nil, nil, nil, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockUser, mockCache)
			}

			err := svc.VerifyEmail(context.Background(), tc.token)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authServiceSuite) TestRefreshToken() {
	token := "refresh-token"
	tokenInfo := domain.TokenInfo{AccessToken: "new-access"}

	tests := []struct {
		name      string
		token     string
		setupMock func(t *mocks.TokenService)
		want      *domain.TokenInfo
		wantErr   error
	}{
		{
			name:  "error",
			token: token,
			setupMock: func(t *mocks.TokenService) {
				t.EXPECT().Refresh(mock.Anything, token).Return(domain.TokenInfo{}, pkg.ErrUnauthorized).Once()
			},
			want:    nil,
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name:  "success",
			token: token,
			setupMock: func(t *mocks.TokenService) {
				t.EXPECT().Refresh(mock.Anything, token).Return(tokenInfo, nil).Once()
			},
			want:    &tokenInfo,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockToken := mocks.NewTokenService(s.T())
			svc := NewAuthService(nil, mockToken, nil, nil, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockToken)
			}

			got, err := svc.RefreshToken(context.Background(), tc.token)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *authServiceSuite) TestIsResetPasswordTokenValid() {
	token := "valid-token"
	email := "test@example.com"

	tests := []struct {
		name      string
		token     string
		setupMock func(c *mocks.CacheStorage)
		want      bool
	}{
		{
			name:  "invalid",
			token: "invalid",
			setupMock: func(c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(pkg.ErrNotFound).Once()
			},
			want: false,
		},
		{
			name:  "valid",
			token: token,
			setupMock: func(c *mocks.CacheStorage) {
				c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
					Run(func(ctx context.Context, key string, dest any) {
						*dest.(*string) = email
					}).Return(nil).Once()
			},
			want: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewAuthService(nil, nil, nil, nil, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockCache)
			}

			got := svc.IsResetPasswordTokenValid(context.Background(), tc.token)
			s.Equal(tc.want, got)
		})
	}
}
