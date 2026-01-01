package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"air-social/internal/domain"
	"air-social/pkg"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserService) VerifyEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserService) UpdatePassword(ctx context.Context, email, newPassword string) error {
	args := m.Called(ctx, email, newPassword)
	return args.Error(0)
}

type MockToken struct {
	mock.Mock
}

func (m *MockToken) CreateSession(ctx context.Context, userID int64, deviceID string) (*domain.TokenInfo, error) {
	args := m.Called(ctx, userID, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenInfo), args.Error(1)
}

func (m *MockToken) Refresh(ctx context.Context, refreshToken string) (*domain.TokenInfo, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenInfo), args.Error(1)
}

func (m *MockToken) RevokeSingle(ctx context.Context, refreshToken string) error {
	return m.Called(ctx, refreshToken).Error(0)
}

func (m *MockToken) Validate(accessToken string) (*jwt.Token, error) {
	args := m.Called(accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockToken) RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error {
	return m.Called(ctx, userID, deviceID).Error(0)
}

func (m *MockToken) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockToken) CleanupDatabase(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Publish(ctx context.Context, topic string, payload any) error {
	return m.Called(ctx, topic, payload).Error(0)
}

func (m *MockQueue) Close() {
	m.Called()
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dst any) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	args := m.Called(ctx, key, val, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) IsExist(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

type MockRoutes struct {
	mock.Mock
}

func (m *MockRoutes) ResetPasswordURL(token string) string {
	return m.Called(token).String(0)
}

func (m *MockRoutes) VerifyEmailURL(token string) string {
	return m.Called(token).String(0)
}

func (m *MockRoutes) Prefix() string {
	return m.Called().String(0)
}

func (m *MockRoutes) SwaggerURL() string {
	return m.Called().String(0)
}

func TestAuthService_Register(t *testing.T) {
	mockUsers := new(MockUserService)
	mockCache := new(MockCache)
	mockToken := new(MockToken)
	mockQueue := new(MockQueue)
	mockRoutes := new(MockRoutes)

	authService := NewAuthService(mockUsers, mockToken, mockQueue, mockRoutes, mockCache)

	validReq := &domain.RegisterRequest{
		Email:    "test@example.com",
		Username: "tester",
		Password: "123456",
	}

	tests := []struct {
		name          string
		input         *domain.RegisterRequest
		setupMocks    func(u *MockUserService)
		expectedError error
	}{
		{
			name:  "success",
			input: validReq,
			setupMocks: func(u *MockUserService) {
				u.On("CreateUser", mock.Anything, mock.MatchedBy(func(input *domain.CreateUserInput) bool {
					return input.Email == validReq.Email &&
						input.Username == validReq.Username &&
						input.PasswordHash != "" && input.PasswordHash != validReq.Password
				})).
					Return(&domain.UserResponse{
						ID:       1,
						Email:    validReq.Email,
						Username: validReq.Username,
					}, nil)

				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockRoutes.On("VerifyEmailURL", mock.Anything).Return("http://test.link")
				mockQueue.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "create user error",
			input: validReq,
			setupMocks: func(u *MockUserService) {
				u.On("CreateUser", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers.ExpectedCalls = nil
			mockUsers.Calls = nil
			mockCache.ExpectedCalls = nil
			mockCache.Calls = nil
			mockQueue.ExpectedCalls = nil
			mockQueue.Calls = nil
			mockRoutes.ExpectedCalls = nil
			mockRoutes.Calls = nil

			tc.setupMocks(mockUsers)
			res, err := authService.Register(context.Background(), tc.input)

			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tc.input.Email, res.Email)
				assert.Equal(t, tc.input.Username, res.Username)
			}
			mockUsers.AssertExpectations(t)
			mockCache.AssertExpectations(t)
			mockQueue.AssertExpectations(t)
			mockRoutes.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockUsers := new(MockUserService)
	mockCache := new(MockCache)
	mockToken := new(MockToken)
	mockQueue := new(MockQueue)
	mockRoutes := new(MockRoutes)

	authService := NewAuthService(mockUsers, mockToken, mockQueue, mockRoutes, mockCache)

	loginReq := &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		DeviceID: "device-123",
	}

	hashedPwd, _ := hashPassword(loginReq.Password)
	user := &domain.User{
		ID:           1,
		Email:        loginReq.Email,
		Username:     "tester",
		PasswordHash: hashedPwd,
		CreatedAt:    time.Now(),
	}

	tokenInfo := &domain.TokenInfo{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	tests := []struct {
		name              string
		input             *domain.LoginRequest
		setupMocks        func()
		expectedUserResp  *domain.UserResponse
		expectedTokenInfo *domain.TokenInfo
		expectedError     error
	}{
		{
			name:  "user not found",
			input: loginReq,
			setupMocks: func() {
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(nil, assert.AnError)
			},
			expectedError: pkg.ErrInvalidCredentials,
		},
		{
			name:  "invalid password",
			input: loginReq,
			setupMocks: func() {
				badUser := *user
				badUser.PasswordHash, _ = hashPassword("wrong-password")
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(&badUser, nil)
			},
			expectedError: pkg.ErrInvalidCredentials,
		},
		{
			name:  "token creation error",
			input: loginReq,
			setupMocks: func() {
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil)
				mockToken.On("CreateSession", mock.Anything, user.ID, loginReq.DeviceID).Return(nil, assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name:  "success",
			input: loginReq,
			setupMocks: func() {
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil)
				mockToken.On("CreateSession", mock.Anything, user.ID, loginReq.DeviceID).Return(tokenInfo, nil)
			},
			expectedUserResp: &domain.UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				Username:  user.Username,
				CreatedAt: user.CreatedAt,
			},
			expectedTokenInfo: tokenInfo,
			expectedError:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers.ExpectedCalls = nil
			mockUsers.Calls = nil
			mockToken.ExpectedCalls = nil
			mockToken.Calls = nil

			tc.setupMocks()

			userResp, tokenInfo, err := authService.Login(context.Background(), tc.input)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedUserResp, userResp)
			assert.Equal(t, tc.expectedTokenInfo, tokenInfo)

			mockUsers.AssertExpectations(t)
			mockToken.AssertExpectations(t)
		})
	}
}
