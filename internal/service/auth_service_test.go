package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"air-social/internal/domain"
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

type MockJWT struct {
	mock.Mock
}

func (m *MockJWT) GenerateAccessToken(userID int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) GenerateRefreshToken(userID int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) Validate(token string) (*jwt.Token, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockUsers := new(MockUserService)
	mockHasher := new(MockHasher)
	mockJWT := new(MockJWT)

	authService := NewAuthService(mockUsers, mockJWT, mockHasher)

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
