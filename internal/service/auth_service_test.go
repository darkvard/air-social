package service

// import (
// 	"context"
// 	"testing"
// 	"time"
//
//

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	"air-social/internal/domain"
// 	"air-social/pkg"
// )

// func TestAuthService_Register(t *testing.T) {
// 	mockUsers := new(mockUserService)
// 	mockCache := new(mockCacheStorage)
// 	mockToken := new(mockTokenService)
// 	mockQueue := new(mockEventQueue)
// 	mockUrl := new(mockURLFactory)

// 	authService := NewAuthService(mockUsers, mockToken, mockUrl, mockQueue, mockCache)

// 	validReq := domain.RegisterRequest{
// 		Email:    "test@example.com",
// 		Username: "tester",
// 		Password: "123456",
// 	}

// 	tests := []struct {
// 		name          string
// 		input         domain.RegisterRequest
// 		setupMocks    func(u *mockUserService)
// 		expectedError error
// 	}{
// 		{
// 			name:  "success",
// 			input: validReq,
// 			setupMocks: func(u *mockUserService) {
// 				u.On("CreateUser", mock.Anything, mock.MatchedBy(func(input domain.CreateUserParams) bool {
// 					return input.Email == validReq.Email &&
// 						input.Username == validReq.Username &&
// 						input.PasswordHashed != "" && input.PasswordHashed != validReq.Password
// 				})).
// 					Return(domain.UserResponse{
// 						ID:       1,
// 						Email:    validReq.Email,
// 						Username: validReq.Username,
// 					}, nil)

// 				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
// 				mockUrl.On("VerifyEmailLink", mock.Anything).Return("http://test.link")
// 				mockQueue.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name:  "create user error",
// 			input: validReq,
// 			setupMocks: func(u *mockUserService) {
// 				u.On("CreateUser", mock.Anything, mock.Anything).Return(domain.UserResponse{}, assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockUsers.ExpectedCalls = nil
// 			mockUsers.Calls = nil
// 			mockCache.ExpectedCalls = nil
// 			mockCache.Calls = nil
// 			mockQueue.ExpectedCalls = nil
// 			mockQueue.Calls = nil
// 			mockUrl.ExpectedCalls = nil
// 			mockUrl.Calls = nil

// 			tc.setupMocks(mockUsers)
// 			res, err := authService.Register(context.Background(), tc.input)

// 			if tc.expectedError != nil {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, res)
// 				assert.Equal(t, tc.input.Email, res.Email)
// 				assert.Equal(t, tc.input.Username, res.Username)
// 			}
// 			mockUsers.AssertExpectations(t)
// 			mockCache.AssertExpectations(t)
// 			mockQueue.AssertExpectations(t)
// 			mockUrl.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_Refresh(t *testing.T) {
// 	mockToken := new(mockTokenService)
// 	authService := NewAuthService(nil, mockToken, nil, nil, nil)

// 	req := domain.RefreshRequest{
// 		RefreshToken: "refresh-token",
// 	}
// 	tokenInfo := domain.TokenInfo{
// 		AccessToken:  "new-access",
// 		RefreshToken: "new-refresh",
// 	}
// 	emptyTokenInfo := domain.TokenInfo{}

// 	tests := []struct {
// 		name          string
// 		setupMocks    func()
// 		expectedResp  domain.TokenInfo
// 		expectedError error
// 	}{
// 		{
// 			name: "success",
// 			setupMocks: func() {
// 				mockToken.On("Refresh", mock.Anything, req.RefreshToken).Return(tokenInfo, nil)
// 			},
// 			expectedResp:  tokenInfo,
// 			expectedError: nil,
// 		},
// 		{
// 			name: "token expired",
// 			setupMocks: func() {
// 				mockToken.On("Refresh", mock.Anything, req.RefreshToken).Return(emptyTokenInfo, pkg.ErrUnauthorized)
// 			},
// 			expectedResp:  emptyTokenInfo,
// 			expectedError: pkg.ErrUnauthorized,
// 		},
// 		{
// 			name: "token revoked",
// 			setupMocks: func() {
// 				mockToken.On("Refresh", mock.Anything, req.RefreshToken).Return(emptyTokenInfo, pkg.ErrUnauthorized)
// 			},
// 			expectedResp:  emptyTokenInfo,
// 			expectedError: pkg.ErrUnauthorized,
// 		},
// 		{
// 			name: "internal error",
// 			setupMocks: func() {
// 				mockToken.On("Refresh", mock.Anything, req.RefreshToken).Return(emptyTokenInfo, assert.AnError)
// 			},
// 			expectedResp:  emptyTokenInfo,
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockToken.ExpectedCalls = nil
// 			mockToken.Calls = nil
// 			tc.setupMocks()

// 			resp, err := authService.RefreshToken(context.Background(), req)

// 			if tc.expectedError != nil {
// 				assert.ErrorIs(t, err, tc.expectedError)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.expectedResp, resp)
// 			}
// 			mockToken.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_Logout(t *testing.T) {
// 	mockToken := new(mockTokenService)
// 	authService := NewAuthService(nil, mockToken, nil, nil, nil)

// 	tests := []struct {
// 		name          string
// 		req           domain.LogoutRequest
// 		setupMocks    func()
// 		expectedError error
// 	}{
// 		{
// 			name: "logout current device",
// 			req: domain.LogoutRequest{
// 				UserID:       1,
// 				DeviceID:     "device-1",
// 				IsAllDevices: false,
// 			},
// 			setupMocks: func() {
// 				mockToken.On("RevokeDeviceSession", mock.Anything, int64(1), "device-1").Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "logout all devices",
// 			req: domain.LogoutRequest{
// 				UserID:       1,
// 				IsAllDevices: true,
// 			},
// 			setupMocks: func() {
// 				mockToken.On("RevokeAllUserSessions", mock.Anything, int64(1)).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "error",
// 			req: domain.LogoutRequest{
// 				UserID:       1,
// 				DeviceID:     "device-1",
// 				IsAllDevices: false,
// 			},
// 			setupMocks: func() {
// 				mockToken.On("RevokeDeviceSession", mock.Anything, int64(1), "device-1").Return(assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockToken.ExpectedCalls = nil
// 			mockToken.Calls = nil
// 			tc.setupMocks()

// 			err := authService.Logout(context.Background(), tc.req)

// 			if tc.expectedError != nil {
// 				assert.ErrorIs(t, err, tc.expectedError)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			mockToken.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_VerifyEmail(t *testing.T) {
// 	mockUsers := new(mockUserService)
// 	mockCache := new(mockCacheStorage)
// 	authService := NewAuthService(mockUsers, nil, nil, nil, mockCache)

// 	token := "verify-token"
// 	email := "test@example.com"

// 	tests := []struct {
// 		name          string
// 		setupMocks    func()
// 		expectedError error
// 	}{
// 		{
// 			name: "success",
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
// 					Run(func(args mock.Arguments) {
// 						arg := args.Get(2).(*string)
// 						*arg = email
// 					}).Return(nil)
// 				mockUsers.On("VerifyEmail", mock.Anything, email).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "token not found in cache",
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 		{
// 			name: "user verify failed",
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
// 					Run(func(args mock.Arguments) {
// 						arg := args.Get(2).(*string)
// 						*arg = email
// 					}).Return(nil)
// 				mockUsers.On("VerifyEmail", mock.Anything, email).Return(assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockUsers.ExpectedCalls = nil
// 			mockUsers.Calls = nil
// 			mockCache.ExpectedCalls = nil
// 			mockCache.Calls = nil
// 			tc.setupMocks()

// 			err := authService.VerifyEmail(context.Background(), token)

// 			if tc.expectedError != nil {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			mockUsers.AssertExpectations(t)
// 			mockCache.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_ForgotPassword(t *testing.T) {
// 	mockUsers := new(mockUserService)
// 	mockCache := new(mockCacheStorage)
// 	mockQueue := new(mockEventQueue)
// 	mockURL := new(mockURLFactory)
// 	authService := NewAuthService(mockUsers, nil, mockURL, mockQueue, mockCache)

// 	req := domain.ForgotPasswordRequest{Email: "test@example.com"}
// 	user := domain.UserResponse{Email: "test@example.com", Username: "test"}

// 	tests := []struct {
// 		name          string
// 		setupMocks    func()
// 		expectedError error
// 	}{
// 		{
// 			name: "success",
// 			setupMocks: func() {
// 				mockUsers.On("GetByEmail", mock.Anything, req.Email).Return(user, nil)
// 				mockCache.On("Set", mock.Anything, mock.Anything, req.Email, mock.Anything).Return(nil)
// 				mockURL.On("ResetPasswordLink", mock.Anything).Return("http://reset.link")
// 				mockQueue.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "user not found",
// 			setupMocks: func() {
// 				mockUsers.On("GetByEmail", mock.Anything, req.Email).Return(domain.UserResponse{}, assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockUsers.ExpectedCalls = nil
// 			mockUsers.Calls = nil
// 			mockCache.ExpectedCalls = nil
// 			mockCache.Calls = nil
// 			mockQueue.ExpectedCalls = nil
// 			mockQueue.Calls = nil
// 			mockURL.ExpectedCalls = nil
// 			mockURL.Calls = nil
// 			tc.setupMocks()

// 			err := authService.ForgotPassword(context.Background(), req)

// 			if tc.expectedError != nil {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			mockUsers.AssertExpectations(t)
// 			mockCache.AssertExpectations(t)
// 			mockQueue.AssertExpectations(t)
// 			mockURL.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_ResetPassword(t *testing.T) {
// 	mockUsers := new(mockUserService)
// 	mockCache := new(mockCacheStorage)
// 	authService := NewAuthService(mockUsers, nil, nil, nil, mockCache)

// 	req := domain.ResetPasswordRequest{
// 		Token:    "reset-token",
// 		Password: "new-password",
// 	}
// 	email := "test@example.com"

// 	tests := []struct {
// 		name             string
// 		isValidateReturn bool
// 		setupMocks       func()
// 		expectedError    error
// 	}{
// 		{
// 			name:             "success update password",
// 			isValidateReturn: false,
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
// 					Run(func(args mock.Arguments) {
// 						arg := args.Get(2).(*string)
// 						*arg = email
// 					}).Return(nil)
// 				mockUsers.On("UpdatePassword", mock.Anything, email, mock.Anything).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name:             "validate only success",
// 			isValidateReturn: true,
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).
// 					Run(func(args mock.Arguments) {
// 						arg := args.Get(2).(*string)
// 						*arg = email
// 					}).Return(nil)
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name:             "token invalid",
// 			isValidateReturn: false,
// 			setupMocks: func() {
// 				mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockUsers.ExpectedCalls = nil
// 			mockUsers.Calls = nil
// 			mockCache.ExpectedCalls = nil
// 			mockCache.Calls = nil
// 			tc.setupMocks()

// 			err := authService.ResetPassword(context.Background(), req, tc.isValidateReturn)

// 			if tc.expectedError != nil {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			mockUsers.AssertExpectations(t)
// 			mockCache.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAuthService_Login(t *testing.T) {
// 	mockUsers := new(mockUserService)
// 	mockCache := new(mockCacheStorage)
// 	mockToken := new(mockTokenService)
// 	mockQueue := new(mockEventQueue)
// 	mockURL := new(mockURLFactory)

// 	authService := NewAuthService(mockUsers, mockToken, mockURL, mockQueue, mockCache)

// 	loginReq := domain.LoginRequest{
// 		Email:    "test@example.com",
// 		Password: "password123",
// 		DeviceID: "device-123",
// 	}

// 	hashedPwd, _ := hashPassword(loginReq.Password)
// 	user := domain.UserResponse{
// 		ID:           1,
// 		Email:        loginReq.Email,
// 		Username:     "tester",
// 		PasswordHash: hashedPwd,
// 		CreatedAt:    time.Now().UTC(),
// 	}

// 	tokenInfo := domain.TokenInfo{
// 		AccessToken:  "access-token",
// 		RefreshToken: "refresh-token",
// 		ExpiresIn:    3600,
// 		TokenType:    "Bearer",
// 	}

// 	tests := []struct {
// 		name              string
// 		input             domain.LoginRequest
// 		setupMocks        func()
// 		expectedUserResp  domain.UserResponse
// 		expectedTokenInfo domain.TokenInfo
// 		expectedError     error
// 	}{
// 		{
// 			name:  "user not found",
// 			input: loginReq,
// 			setupMocks: func() {
// 				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(domain.UserResponse{}, pkg.ErrNotFound)
// 			},
// 			expectedError: pkg.ErrInvalidCredentials,
// 		},
// 		{
// 			name:  "invalid password",
// 			input: loginReq,
// 			setupMocks: func() {
// 				badUser := user
// 				badUser.PasswordHash, _ = hashPassword("wrong-password")
// 				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(badUser, nil)
// 			},
// 			expectedError: pkg.ErrInvalidCredentials,
// 		},
// 		{
// 			name:  "token creation error",
// 			input: loginReq,
// 			setupMocks: func() {
// 				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil)
// 				mockToken.On("CreateSession", mock.Anything, user.ID, loginReq.DeviceID).Return(domain.TokenInfo{}, assert.AnError)
// 			},
// 			expectedError: assert.AnError,
// 		},
// 		{
// 			name:  "success",
// 			input: loginReq,
// 			setupMocks: func() {
// 				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil)
// 				mockToken.On("CreateSession", mock.Anything, user.ID, loginReq.DeviceID).Return(tokenInfo, nil)
// 			},
// 			expectedUserResp: domain.UserResponse{
// 				ID:           user.ID,
// 				Email:        user.Email,
// 				Username:     user.Username,
// 				CreatedAt:    user.CreatedAt,
// 				PasswordHash: user.PasswordHash,
// 			},
// 			expectedTokenInfo: tokenInfo,
// 			expectedError:     nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			mockUsers.ExpectedCalls = nil
// 			mockUsers.Calls = nil
// 			mockToken.ExpectedCalls = nil
// 			mockToken.Calls = nil

// 			tc.setupMocks()

// 			userResp, tokenInfo, err := authService.Login(context.Background(), tc.input)

// 			assert.ErrorIs(t, err, tc.expectedError)
// 			assert.Equal(t, tc.expectedUserResp, userResp)
// 			assert.Equal(t, tc.expectedTokenInfo, tokenInfo)

// 			mockUsers.AssertExpectations(t)
// 			mockToken.AssertExpectations(t)
// 		})
// 	}
// }
