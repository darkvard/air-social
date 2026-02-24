package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/auth"
	"air-social/internal/domain/auth/token"
	tokenmocks "air-social/internal/domain/auth/token/mocks"
	verifymocks "air-social/internal/domain/auth/verify/mocks"
	"air-social/internal/domain/user"
	usermocks "air-social/internal/domain/user/mocks"
	"air-social/pkg"
)

type authUseCaseSuite struct {
	suite.Suite
}

func TestAuthUseCaseSuite(t *testing.T) {
	suite.Run(t, new(authUseCaseSuite))
}

func (s *authUseCaseSuite) TestRegister() {
	var (
		email    = "test@example.com"
		username = "testuser"
		password = "password"
	)

	params := auth.RegisterParams{
		Email:    email,
		Username: username,
		Password: password,
	}

	createdUser := &user.User{
		ID:       1,
		Email:    email,
		Username: username,
	}

	type testDeps struct {
		userAccount *usermocks.MockAccountUseCase
		verify      *verifymocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantUser  *user.User
		wantErr   error
	}{
		{
			name: "create_user_error",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					CreateUser(mock.Anything, mock.MatchedBy(func(p user.CreateParams) bool {
						return p.Email == email && p.Username == username && p.NewPassword == password
					})).
					Return(nil, pkg.ErrAlreadyExists).Once()
			},
			wantUser: nil,
			wantErr:  pkg.ErrAlreadyExists,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					CreateUser(mock.Anything, mock.Anything).
					Return(createdUser, nil).Once()

				deps.verify.EXPECT().
					SendVerification(mock.Anything, email, username).
					Return(nil).Once()
			},
			wantUser: createdUser,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccount := usermocks.NewMockAccountUseCase(s.T())
			mockVerify := verifymocks.NewMockProvider(s.T())

			uc := auth.NewUseCase(auth.Deps{
				UserAccount:    mockAccount,
				VerifyProvider: mockVerify,
				TokenProvider:  nil, // Not used in Register
				TokenRepo:      nil, // Not used in Register
				UserFetch:      nil, // Not used in Register
				Cache:          nil, // Not used in Register
			})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{userAccount: mockAccount, verify: mockVerify})
			}

			got, err := uc.Register(context.Background(), params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.wantUser, got)
			}
		})
	}
}

func (s *authUseCaseSuite) TestLogin() {
	var (
		email    = "test@example.com"
		password = "password"
		deviceID = "device-1"
		userID   = int64(1)
	)

	params := auth.LoginParams{
		Email:    email,
		Password: password,
		DeviceID: deviceID,
	}

	authUser := &user.User{ID: userID, Email: email}
	accessToken := token.AccessTokenResult{Token: "access", ExpiresAt: time.Now()}
	refreshToken := token.RefreshTokenResult{Raw: "refresh", Hashed: "hashed", ExpiresAt: time.Now()}

	type testDeps struct {
		userAccount   *usermocks.MockAccountUseCase
		tokenProvider *tokenmocks.MockProvider
		tokenRepo     *tokenmocks.MockRepository
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "invalid_credentials",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					Authenticate(mock.Anything, user.AuthenticateParams{Email: email, Password: password}).
					Return(nil, pkg.ErrInvalidCredentials).Once()
			},
			wantErr: pkg.ErrInvalidCredentials, // OrInternalError(ErrInvalidCredentials, ErrInvalidCredentials) -> ErrInvalidCredentials
		},
		{
			name: "access_token_error",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					Authenticate(mock.Anything, mock.Anything).
					Return(authUser, nil).Once()

				deps.tokenProvider.EXPECT().
					GenerateRefreshToken().
					Return(refreshToken).Once()

				deps.tokenProvider.EXPECT().
					GenerateAccessToken(userID, deviceID).
					Return(token.AccessTokenResult{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "repo_create_error",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					Authenticate(mock.Anything, mock.Anything).
					Return(authUser, nil).Once()

				deps.tokenProvider.EXPECT().
					GenerateRefreshToken().
					Return(refreshToken).Once()

				deps.tokenProvider.EXPECT().
					GenerateAccessToken(userID, deviceID).
					Return(accessToken, nil).Once()

				deps.tokenRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.userAccount.EXPECT().
					Authenticate(mock.Anything, mock.Anything).
					Return(authUser, nil).Once()

				deps.tokenProvider.EXPECT().
					GenerateRefreshToken().
					Return(refreshToken).Once()

				deps.tokenProvider.EXPECT().
					GenerateAccessToken(userID, deviceID).
					Return(accessToken, nil).Once()

				deps.tokenRepo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(t *token.RefreshToken) bool {
						return t.UserID == userID && t.TokenHash == refreshToken.Hashed
					})).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockAccount := usermocks.NewMockAccountUseCase(s.T())
			mockTokenProv := tokenmocks.NewMockProvider(s.T())
			mockTokenRepo := tokenmocks.NewMockRepository(s.T())

			uc := auth.NewUseCase(auth.Deps{
				UserAccount:   mockAccount,
				TokenProvider: mockTokenProv,
				TokenRepo:     mockTokenRepo,
			})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{
					userAccount:   mockAccount,
					tokenProvider: mockTokenProv,
					tokenRepo:     mockTokenRepo,
				})
			}

			_, _, err := uc.Login(context.Background(), params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authUseCaseSuite) TestLogout() {
	var (
		userID   = int64(1)
		deviceID = "device-1"
		tokenStr = "access-token"
	)

	type testDeps struct {
		tokenProvider *tokenmocks.MockProvider
		tokenRepo     *tokenmocks.MockRepository
	}

	tests := []struct {
		name      string
		args      auth.LogoutParams
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "already_blacklisted",
			args: auth.LogoutParams{UserID: userID, DeviceID: deviceID, Token: tokenStr},
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().
					IsBlacklisted(mock.Anything, tokenStr).
					Return(true).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "revoke_device_error",
			args: auth.LogoutParams{UserID: userID, DeviceID: deviceID, Token: tokenStr, IsAllDevices: false},
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().IsBlacklisted(mock.Anything, tokenStr).Return(false).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "revoke_all_error",
			args: auth.LogoutParams{UserID: userID, DeviceID: deviceID, Token: tokenStr, IsAllDevices: true},
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().IsBlacklisted(mock.Anything, tokenStr).Return(false).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByUser(mock.Anything, userID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_single_device",
			args: auth.LogoutParams{UserID: userID, DeviceID: deviceID, Token: tokenStr, IsAllDevices: false},
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().IsBlacklisted(mock.Anything, tokenStr).Return(false).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(nil).Once()
				deps.tokenProvider.EXPECT().AddToBlacklist(mock.Anything, tokenStr, mock.Anything).Return().Once()
			},
			wantErr: nil,
		},
		{
			name: "success_all_devices",
			args: auth.LogoutParams{UserID: userID, DeviceID: deviceID, Token: tokenStr, IsAllDevices: true},
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().IsBlacklisted(mock.Anything, tokenStr).Return(false).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByUser(mock.Anything, userID).Return(nil).Once()
				deps.tokenProvider.EXPECT().AddToBlacklist(mock.Anything, tokenStr, mock.Anything).Return().Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockTokenProv := tokenmocks.NewMockProvider(s.T())
			mockTokenRepo := tokenmocks.NewMockRepository(s.T())
			uc := auth.NewUseCase(auth.Deps{TokenProvider: mockTokenProv, TokenRepo: mockTokenRepo})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{tokenProvider: mockTokenProv, tokenRepo: mockTokenRepo})
			}
			err := uc.Logout(context.Background(), tc.args)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authUseCaseSuite) TestForgotPassword() {
	var (
		email    = "test@example.com"
		username = "testuser"
	)

	userObj := &user.User{
		Email:    email,
		Username: username,
	}

	type testDeps struct {
		userFetch *usermocks.MockFetchUseCase
		verify    *verifymocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "user_not_found",
			setupMock: func(deps testDeps) {
				deps.userFetch.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.userFetch.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(userObj, nil).Once()
				deps.verify.EXPECT().
					SendPasswordReset(mock.Anything, email, username).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockFetch := usermocks.NewMockFetchUseCase(s.T())
			mockVerify := verifymocks.NewMockProvider(s.T())

			uc := auth.NewUseCase(auth.Deps{
				UserFetch:      mockFetch,
				VerifyProvider: mockVerify,
			})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{userFetch: mockFetch, verify: mockVerify})
			}

			err := uc.ForgotPassword(context.Background(), email)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authUseCaseSuite) TestResetPassword() {
	var (
		emailToken = "valid-token"
		email      = "test@example.com"
		password   = "new-password"
	)

	params := auth.ResetPasswordParams{
		EmailToken: emailToken,
		Password:   password,
	}

	type testDeps struct {
		verify      *verifymocks.MockProvider
		userAccount *usermocks.MockAccountUseCase
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "verify_token_error",
			setupMock: func(deps testDeps) {
				deps.verify.EXPECT().
					VerifyPasswordReset(mock.Anything, emailToken).
					Return("", pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "update_password_error",
			setupMock: func(deps testDeps) {
				deps.verify.EXPECT().
					VerifyPasswordReset(mock.Anything, emailToken).
					Return(email, nil).Once()
				deps.userAccount.EXPECT().
					ResetPassword(mock.Anything, user.ResetPasswordParams{
						Email:       email,
						NewPassword: password,
					}).
					Return(pkg.ErrInvalidCredentials).Once()
			},
			wantErr: pkg.ErrInvalidCredentials,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.verify.EXPECT().
					VerifyPasswordReset(mock.Anything, emailToken).
					Return(email, nil).Once()
				deps.userAccount.EXPECT().
					ResetPassword(mock.Anything, user.ResetPasswordParams{
						Email:       email,
						NewPassword: password,
					}).
					Return(nil).Once()
				deps.verify.EXPECT().
					InvalidatePasswordReset(mock.Anything, emailToken).
					Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockVerify := verifymocks.NewMockProvider(s.T())
			mockAccount := usermocks.NewMockAccountUseCase(s.T())

			uc := auth.NewUseCase(auth.Deps{
				VerifyProvider: mockVerify,
				UserAccount:    mockAccount,
			})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{verify: mockVerify, userAccount: mockAccount})
			}

			err := uc.ResetPassword(context.Background(), params)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *authUseCaseSuite) TestValidateResetPasswordToken() {
	var tokenStr = "token"

	type testDeps struct {
		verify *verifymocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		want      bool
	}{
		{
			name: "valid",
			setupMock: func(deps testDeps) {
				deps.verify.EXPECT().
					ValidateResetPasswordToken(mock.Anything, tokenStr).
					Return(true).Once()
			},
			want: true,
		},
		{
			name: "invalid",
			setupMock: func(deps testDeps) {
				deps.verify.EXPECT().
					ValidateResetPasswordToken(mock.Anything, tokenStr).
					Return(false).Once()
			},
			want: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockVerify := verifymocks.NewMockProvider(s.T())
			uc := auth.NewUseCase(auth.Deps{VerifyProvider: mockVerify})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{verify: mockVerify})
			}

			got := uc.ValidateResetPasswordToken(context.Background(), tokenStr)
			s.Equal(tc.want, got)
		})
	}
}

func (s *authUseCaseSuite) TestRefreshToken() {
	var (
		rawRefreshToken = "raw-refresh-token"
		hashedToken     = "hashed-token"
		userID          = int64(1)
		deviceID        = "device-1"
		tokenID         = int64(100)
	)

	storedToken := &token.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		DeviceID:  deviceID,
		TokenHash: hashedToken,
	}

	newAccessToken := token.AccessTokenResult{Token: "new-access", ExpiresAt: time.Now()}
	newRefreshToken := token.RefreshTokenResult{Raw: "new-refresh", Hashed: "new-hashed", ExpiresAt: time.Now()}

	type testDeps struct {
		tokenRepo     *tokenmocks.MockRepository
		tokenProvider *tokenmocks.MockProvider
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "token_not_found",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "verify_failed_revoked",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(storedToken, nil).Once()
				deps.tokenProvider.EXPECT().VerifyRefreshToken(*storedToken).Return(false, pkg.ErrTokenRevoked).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByUser(mock.Anything, userID).Return(nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "verify_failed_invalid",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(storedToken, nil).Once()
				deps.tokenProvider.EXPECT().VerifyRefreshToken(*storedToken).Return(false, nil).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "rotate_revoke_error",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(storedToken, nil).Once()
				deps.tokenProvider.EXPECT().VerifyRefreshToken(*storedToken).Return(true, nil).Once()
				deps.tokenRepo.EXPECT().UpdateRevoked(mock.Anything, tokenID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "rotate_access_token_error",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(storedToken, nil).Once()
				deps.tokenProvider.EXPECT().VerifyRefreshToken(*storedToken).Return(true, nil).Once()
				deps.tokenRepo.EXPECT().UpdateRevoked(mock.Anything, tokenID).Return(nil).Once()
				deps.tokenProvider.EXPECT().GenerateRefreshToken().Return(newRefreshToken).Once()
				deps.tokenProvider.EXPECT().GenerateAccessToken(userID, deviceID).Return(token.AccessTokenResult{}, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().HashToken(rawRefreshToken).Return(hashedToken).Once()
				deps.tokenRepo.EXPECT().GetByHash(mock.Anything, hashedToken).Return(storedToken, nil).Once()
				deps.tokenProvider.EXPECT().VerifyRefreshToken(*storedToken).Return(true, nil).Once()
				deps.tokenRepo.EXPECT().UpdateRevoked(mock.Anything, tokenID).Return(nil).Once()
				deps.tokenProvider.EXPECT().GenerateRefreshToken().Return(newRefreshToken).Once()
				deps.tokenProvider.EXPECT().GenerateAccessToken(userID, deviceID).Return(newAccessToken, nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockTokenRepo := tokenmocks.NewMockRepository(s.T())
			mockTokenProv := tokenmocks.NewMockProvider(s.T())

			uc := auth.NewUseCase(auth.Deps{
				TokenRepo:     mockTokenRepo,
				TokenProvider: mockTokenProv,
			})

			if tc.setupMock != nil {
				tc.setupMock(testDeps{tokenRepo: mockTokenRepo, tokenProvider: mockTokenProv})
			}

			_, err := uc.RefreshToken(context.Background(), rawRefreshToken)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
