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
		setupMock func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager, verifySvc *mocks.VerifyService, cache *mocks.CacheStorage)
		want      *domain.User
		wantErr   error
	}{
		{
			name: "create_user_error",
			args: args{input: input},
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager, verifySvc *mocks.VerifyService, cache *mocks.CacheStorage) {
				accountMgr.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil, pkg.ErrAlreadyExists).Once()
			},
			want:    nil,
			wantErr: pkg.ErrAlreadyExists,
		},
		{
			name: "success",
			args: args{input: input},
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager, verifySvc *mocks.VerifyService, cache *mocks.CacheStorage) {
				accountMgr.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(p domain.CreateUserParams) bool {
					return p.Email == input.Email && p.Username == input.Username && p.PasswordHashed != ""
				})).Return(userResp, nil).Once()

				verifySvc.EXPECT().SendEmailVerification(mock.Anything, input.Email, input.Username).Return(nil).Once()
			},
			want:    userResp,
			wantErr: nil,
		},
		{
			name: "send_email_error_ignored",
			args: args{input: input},
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager, verifySvc *mocks.VerifyService, cache *mocks.CacheStorage) {
				accountMgr.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(userResp, nil).Once()
				verifySvc.EXPECT().SendEmailVerification(mock.Anything, input.Email, input.Username).Return(assert.AnError).Once()
			},
			want:    userResp,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccountMgr := mocks.NewUserAccountManager(s.T())
			mockSessionMgr := mocks.NewSessionManager(s.T())
			mockVerify := mocks.NewVerifyService(s.T())
			mockCache := mocks.NewCacheStorage(s.T())

			svc := NewAuthService(mockAccountMgr, mockSessionMgr, mockVerify, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockAccountMgr, mockSessionMgr, mockVerify, mockCache)
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
	hashedPwd, _ := pkg.HashPassword(password)

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
		setupMock func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager)
		wantUser  *domain.User
		wantToken domain.TokenInfo
		wantErr   error
	}{
		{
			name:  "user_not_found",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, input.Email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name:  "user_repo_error",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, input.Email).Return(nil, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
		{
			name:  "invalid_password",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager) {
				otherHash, _ := pkg.HashPassword("other")
				accountMgr.EXPECT().GetByEmail(mock.Anything, input.Email).Return(&domain.User{PasswordHash: otherHash}, nil).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name:  "token_creation_error",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, input.Email).Return(user, nil).Once()
				sessionMgr.EXPECT().CreateSession(mock.Anything, user.ID, input.DeviceID).Return(domain.TokenInfo{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:  "success",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, sessionMgr *mocks.SessionManager) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, input.Email).Return(user, nil).Once()
				sessionMgr.EXPECT().CreateSession(mock.Anything, user.ID, input.DeviceID).Return(tokenInfo, nil).Once()
			},
			wantUser:  user,
			wantToken: tokenInfo,
			wantErr:   nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccountMgr := mocks.NewUserAccountManager(s.T())
			mockSessionMgr := mocks.NewSessionManager(s.T())
			svc := NewAuthService(mockAccountMgr, mockSessionMgr, nil, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockAccountMgr, mockSessionMgr)
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
		setupMock func(sessionMgr *mocks.SessionManager, cache *mocks.CacheStorage)
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
			setupMock: func(sessionMgr *mocks.SessionManager, cache *mocks.CacheStorage) {
				cache.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(true, nil).Once()
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
			setupMock: func(sessionMgr *mocks.SessionManager, cache *mocks.CacheStorage) {
				cache.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				sessionMgr.EXPECT().RevokeAllUserSessions(mock.Anything, userID).Return(nil).Once()
				cache.EXPECT().Set(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken), "revoked", mock.Anything).Return(nil).Once()
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
			setupMock: func(sessionMgr *mocks.SessionManager, cache *mocks.CacheStorage) {
				cache.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				sessionMgr.EXPECT().RevokeDeviceSession(mock.Anything, userID, deviceID).Return(nil).Once()
				cache.EXPECT().Set(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken), "revoked", mock.Anything).Return(nil).Once()
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
			setupMock: func(sessionMgr *mocks.SessionManager, cache *mocks.CacheStorage) {
				cache.EXPECT().IsExist(mock.Anything, domain.GetBlacklistTokenKey(tokenMeta.AccessToken)).Return(false, nil).Once()
				sessionMgr.EXPECT().RevokeAllUserSessions(mock.Anything, userID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockSessionMgr := mocks.NewSessionManager(s.T())
			mockCache := mocks.NewCacheStorage(s.T())
			svc := NewAuthService(nil, mockSessionMgr, nil, mockCache)

			if tc.setupMock != nil {
				tc.setupMock(mockSessionMgr, mockCache)
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
		setupMock func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService)
		wantErr   error
	}{
		{
			name:  "user_not_found",
			email: email,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, email).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "success",
			email: email,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				accountMgr.EXPECT().GetByEmail(mock.Anything, email).Return(user, nil).Once()
				verifySvc.EXPECT().SendPasswordReset(mock.Anything, email, user.Username).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccountMgr := mocks.NewUserAccountManager(s.T())
			mockVerify := mocks.NewVerifyService(s.T())

			svc := NewAuthService(mockAccountMgr, nil, mockVerify, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockAccountMgr, mockVerify)
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
		setupMock func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService)
		wantErr   error
	}{
		{
			name:  "token_invalid",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyPasswordResetToken(mock.Anything, token).Return("", pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "success",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyPasswordResetToken(mock.Anything, token).Return(email, nil).Once()

				accountMgr.EXPECT().UpdatePassword(mock.Anything, email, mock.Anything).Return(nil).Once()
				verifySvc.EXPECT().InvalidatePasswordResetToken(mock.Anything, token).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:  "update_password_error",
			input: input,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyPasswordResetToken(mock.Anything, token).Return(email, nil).Once()
				accountMgr.EXPECT().UpdatePassword(mock.Anything, email, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccountMgr := mocks.NewUserAccountManager(s.T())
			mockVerify := mocks.NewVerifyService(s.T())

			svc := NewAuthService(mockAccountMgr, nil, mockVerify, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockAccountMgr, mockVerify)
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
		setupMock func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService)
		wantErr   error
	}{
		{
			name:  "token_invalid",
			token: token,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyEmailToken(mock.Anything, token).Return("", pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name:  "success",
			token: token,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyEmailToken(mock.Anything, token).Return(email, nil).Once()

				accountMgr.EXPECT().VerifyEmail(mock.Anything, email).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:  "user_verify_error",
			token: token,
			setupMock: func(accountMgr *mocks.UserAccountManager, verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyEmailToken(mock.Anything, token).Return(email, nil).Once()
				accountMgr.EXPECT().VerifyEmail(mock.Anything, email).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccountMgr := mocks.NewUserAccountManager(s.T())
			mockVerify := mocks.NewVerifyService(s.T())

			svc := NewAuthService(mockAccountMgr, nil, mockVerify, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockAccountMgr, mockVerify)
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
		setupMock func(sessionMgr *mocks.SessionManager)
		want      domain.TokenInfo
		wantErr   error
	}{
		{
			name:  "error",
			token: token,
			setupMock: func(sessionMgr *mocks.SessionManager) {
				sessionMgr.EXPECT().Refresh(mock.Anything, token).Return(domain.TokenInfo{}, pkg.ErrUnauthorized).Once()
			},
			want:    domain.TokenInfo{},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name:  "success",
			token: token,
			setupMock: func(sessionMgr *mocks.SessionManager) {
				sessionMgr.EXPECT().Refresh(mock.Anything, token).Return(tokenInfo, nil).Once()
			},
			want:    tokenInfo,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockSessionMgr := mocks.NewSessionManager(s.T())
			svc := NewAuthService(nil, mockSessionMgr, nil, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockSessionMgr)
			}

			got, err := svc.RefreshToken(context.Background(), tc.token)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.AccessToken)
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
		setupMock func(verifySvc *mocks.VerifyService)
		want      bool
	}{
		{
			name:  "invalid",
			token: "invalid",
			setupMock: func(verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyPasswordResetToken(mock.Anything, "invalid").Return("", pkg.ErrNotFound).Once()
			},
			want: false,
		},
		{
			name:  "valid",
			token: token,
			setupMock: func(verifySvc *mocks.VerifyService) {
				verifySvc.EXPECT().VerifyPasswordResetToken(mock.Anything, token).Return(email, nil).Once()
			},
			want: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockVerify := mocks.NewVerifyService(s.T())
			svc := NewAuthService(nil, nil, mockVerify, nil)

			if tc.setupMock != nil {
				tc.setupMock(mockVerify)
			}

			got := svc.IsResetPasswordTokenValid(context.Background(), tc.token)
			s.Equal(tc.want, got)
		})
	}
}
