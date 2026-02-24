package auth_test

import (
	"context"
	"testing"
	"time"

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

	params := auth.LogoutParams{
		UserID:   userID,
		DeviceID: deviceID,
		Token:    tokenStr,
	}

	type testDeps struct {
		tokenProvider *tokenmocks.MockProvider
		tokenRepo     *tokenmocks.MockRepository
	}

	tests := []struct {
		name      string
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "already_blacklisted",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().
					IsBlacklisted(mock.Anything, tokenStr).
					Return(true).Once()
			},
			wantErr: pkg.ErrUnauthorized,
		},
		{
			name: "success",
			setupMock: func(deps testDeps) {
				deps.tokenProvider.EXPECT().IsBlacklisted(mock.Anything, tokenStr).Return(false).Once()
				deps.tokenRepo.EXPECT().UpdateRevokedByDevice(mock.Anything, userID, deviceID).Return(nil).Once()
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
			err := uc.Logout(context.Background(), params)
			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
