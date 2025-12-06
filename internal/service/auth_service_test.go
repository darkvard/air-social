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

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockHasher) Verify(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
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

func (m *MockToken) Refresh(ctx context.Context, raw string) (*domain.TokenInfo, error) {
	args := m.Called(ctx, raw)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenInfo), args.Error(1)
}

func (m *MockToken) RevokeSingle(ctx context.Context, raw string) error {
	return m.Called(ctx, raw).Error(0)
}

func (m *MockToken) Validate(access string) (*jwt.Token, error) {
	args := m.Called(access)
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

func TestAuthService_Register(t *testing.T) {
	mockUsers := new(MockUserService)
	mockHasher := new(MockHasher)
	mockToken := new(MockToken)

	authService := NewAuthService(mockUsers, mockToken, mockHasher)

	validReq := &domain.RegisterRequest{
		Email:    "test@example.com",
		Username: "tester",
		Password: "123456",
	}
	hashed := "hashed-pass"

	tests := []struct {
		name          string
		input         *domain.RegisterRequest
		setupMocks    func(u *MockUserService, h *MockHasher)
		expectedError error
	}{
		{
			name:  "hash error",
			input: validReq,
			setupMocks: func(u *MockUserService, h *MockHasher) {
				h.On("Hash", validReq.Password).Return("", assert.AnError)
			},
			expectedError: assert.AnError,
		},
		{
			name:  "success",
			input: validReq,
			setupMocks: func(u *MockUserService, h *MockHasher) {
				h.On("Hash", validReq.Password).Return(hashed, nil)

				expectedCreateInput := &domain.CreateUserInput{
					Email:        validReq.Email,
					Username:     validReq.Username,
					PasswordHash: hashed,
				}

				u.On("CreateUser", mock.Anything, expectedCreateInput).
					Return(&domain.UserResponse{
						ID:       1,
						Email:    validReq.Email,
						Username: validReq.Username,
					}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers.ExpectedCalls = nil
			mockUsers.Calls = nil
			mockHasher.ExpectedCalls = nil
			mockHasher.Calls = nil

			tc.setupMocks(mockUsers, mockHasher)
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
			mockHasher.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockUsers := new(MockUserService)
	mockHasher := new(MockHasher)
	mockToken := new(MockToken)

	authService := NewAuthService(mockUsers, mockToken, mockHasher)

	loginReq := &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		DeviceID: "device-123",
	}

	user := &domain.User{
		ID:           1,
		Email:        loginReq.Email,
		Username:     "tester",
		PasswordHash: "hashed-password",
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
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil) 
				mockHasher.On("Verify", loginReq.Password, user.PasswordHash).Return(false) 
			},
			expectedError: pkg.ErrInvalidCredentials,
		},
		{
			name:  "token creation error",
			input: loginReq,
			setupMocks: func() {
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil) 
				mockHasher.On("Verify", loginReq.Password, user.PasswordHash).Return(true) 
				mockToken.On("CreateSession", mock.Anything, user.ID, loginReq.DeviceID).Return(nil, assert.AnError) 
			},
			expectedError: assert.AnError,
		},
		{
			name:  "success",
			input: loginReq,
			setupMocks: func() {
				mockUsers.On("GetByEmail", mock.Anything, loginReq.Email).Return(user, nil)
				mockHasher.On("Verify", loginReq.Password, user.PasswordHash).Return(true)
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
			mockHasher.ExpectedCalls = nil
			mockHasher.Calls = nil
			mockToken.ExpectedCalls = nil
			mockToken.Calls = nil

			tc.setupMocks()

			userResp, tokenInfo, err := authService.Login(context.Background(), tc.input)

			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedUserResp, userResp)
			assert.Equal(t, tc.expectedTokenInfo, tokenInfo)

			mockUsers.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
			mockToken.AssertExpectations(t)
		})
	}
}
