package service

import (
	"context"
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

type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) GetPresignedURL(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(domain.PresignedFileResponse), args.Error(1)

}

func (m *MockMediaService) ConfirmUpload(ctx context.Context, objectName string, userID int64) (string, error) {
	args := m.Called(ctx, objectName, userID)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockMediaService) DeleteFile(ctx context.Context, fullURL string) error {
	return m.Called(ctx, fullURL).Error(0)
}

func (m *MockMediaService) GetPublicURL(objectName string) string {
	return m.Called(objectName).Get(0).(string)
}

func TestUserService_Create(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockMedia := new(MockMediaService)
	service := NewUserService(mockRepo, mockMedia)

	input := domain.CreateUserRequest{
		Email:        "email@example.com",
		Username:     "test",
		PasswordHash: "hash",
	}

	tests := []struct {
		name          string
		input         domain.CreateUserRequest
		setupMocks    func(m *MockUserRepo, media *MockMediaService)
		expectedError error
	}{
		{
			name:  "email already exists",
			input: input,
			setupMocks: func(m *MockUserRepo, media *MockMediaService) {
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
			setupMocks: func(m *MockUserRepo, media *MockMediaService) {
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
				media.On("GetPublicURL", mock.Anything).Return("")
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks(mockRepo, mockMedia)

			u, err := service.CreateUser(context.Background(), tc.input)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NotNil(t, u)
			}

			mockRepo.AssertExpectations(t)
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			mockMedia.AssertExpectations(t)
			mockMedia.ExpectedCalls = nil
			mockMedia.Calls = nil
		})
	}

}
