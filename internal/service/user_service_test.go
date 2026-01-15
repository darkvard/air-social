package service

import (
	"context"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"air-social/internal/domain"
	"air-social/pkg"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)

}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) UpdateProfileImages(ctx context.Context, userID int64, url string, imageType domain.FileType) error {
	return m.Called(ctx, userID, imageType).Error(0)
}

type MockFile struct {
	mock.Mock
}

func (m *MockFile) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	args := m.Called(ctx, file, header, folder)
	return args.String(0), args.Error(1)
}

func (m *MockFile) DeleteFile(ctx context.Context, path string) error {
	return m.Called(ctx, path).Error(0)
}

func TestUserService_Create(t *testing.T) {
	mockRepo := new(MockUserRepo)
	service := NewUserService(mockRepo, nil)

	input := &domain.CreateUserInput{
		Email:        "email@example.com",
		Username:     "test",
		PasswordHash: "hash",
	}

	tests := []struct {
		name          string
		input         *domain.CreateUserInput
		setupMocks    func(m *MockUserRepo)
		expectedError error
	}{
		{
			name:  "email already exists",
			input: input,
			setupMocks: func(m *MockUserRepo) {
				m.On("GetByEmail", mock.Anything, input.Email).Return(
					&domain.User{
						Email:        input.Email,
						Username:     input.Username,
						PasswordHash: input.PasswordHash,
					}, nil)
			},
			expectedError: pkg.ErrAlreadyExists,
		},
		{
			name:  "successfully created",
			input: input,
			setupMocks: func(m *MockUserRepo) {
				m.On("GetByEmail", mock.Anything, input.Email).Return(nil, nil)
				m.On("Create",
					mock.Anything,
					mock.MatchedBy(func(u *domain.User) bool {
						return u.Email == input.Email &&
							u.Username == input.Username &&
							u.PasswordHash == input.PasswordHash
					}),
				).Run(func(args mock.Arguments) {
					u := args.Get(1).(*domain.User)
					u.ID = 123
					u.CreatedAt = time.Now().UTC()
				}).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(mockRepo)

			_, err := service.CreateUser(context.Background(), tc.input)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
		})
	}

}
