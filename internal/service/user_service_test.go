package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain"
	"air-social/internal/mocks"
	"air-social/pkg"
)

type userServiceSuite struct {
	suite.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(userServiceSuite))
}

func (s *userServiceSuite) TestCreateUser() {
	input := domain.CreateUserParams{
		Email:          "email@example.com",
		Username:       "test",
		PasswordHashed: "hash",
	}

	tests := []struct {
		name      string
		setupMock func(repo *mocks.UserRepository, media *mocks.MediaService)
		wantErr   error
	}{
		{
			name: "repo_get_error",
			setupMock: func(repo *mocks.UserRepository, media *mocks.MediaService) {
				repo.EXPECT().
					GetByEmail(mock.Anything, input.Email).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "repo_get_exists",
			setupMock: func(repo *mocks.UserRepository, media *mocks.MediaService) {
				foundUser := &domain.User{Email: input.Email}

				repo.EXPECT().
					GetByEmail(mock.Anything, input.Email).
					Return(foundUser, nil).
					Once()
			},
			wantErr: pkg.ErrAlreadyExists,
		},
		{
			name: "success",
			setupMock: func(repo *mocks.UserRepository, media *mocks.MediaService) {
				repo.EXPECT().
					GetByEmail(mock.Anything, input.Email).
					Return(nil, pkg.ErrNotFound).
					Once()

				repo.EXPECT().
					Create(
						mock.Anything,
						mock.MatchedBy(func(u *domain.User) bool {
							return u.Email == input.Email &&
								u.Username == input.Username &&
								u.PasswordHash == input.PasswordHashed
						}),
					).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := mocks.NewUserRepository(s.T())
			mockMedia := mocks.NewMediaService(s.T())

			tc.setupMock(mockRepo, mockMedia)

			userSvc := NewUserService(mockRepo, mockMedia)

			got, err := userSvc.CreateUser(context.Background(), input)

			if tc.wantErr != nil {
				if tc.wantErr != assert.AnError {
					s.ErrorIs(err, tc.wantErr)
				} else {
					s.Error(err)
				}
			} else {
				s.NoError(err)
				s.Equal(input.Email, got.Email)
				s.Equal(input.Username, got.Username)
			}

		})
	}
}

// todo: implement tests, refactor old test