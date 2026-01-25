package service

// import (
// 	"context"
// 	"testing"
//
//

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"

// 	"air-social/internal/domain"
// 	"air-social/pkg"
// )

// type userServiceSuite struct {
// 	suite.Suite
// 	userRepo *mockUserRepository
// 	mediaSvc *mockMediaService
// 	userSvc  UserService
// }

// func (s *userServiceSuite) SetupTest() {
// 	s.userRepo = new(mockUserRepository)
// 	s.mediaSvc = new(mockMediaService)
// 	s.userSvc = NewUserService(s.userRepo, s.mediaSvc)
// }

// func TestUserServiceSuite(t *testing.T) {
// 	suite.Run(t, new(userServiceSuite))
// }

// func (s *userServiceSuite) TestCreateUser() {
// 	input := domain.CreateUserParams{
// 		Email:          "email@example.com",
// 		Username:       "test",
// 		PasswordHashed: "hash",
// 	}

// 	tests := []struct {
// 		name    string
// 		mock    func()
// 		wantErr error
// 	}{
// 		{
// 			name: "repo_get_error",
// 			mock: func() {
// 				s.userRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, assert.AnError).Once()
// 			},
// 			wantErr: pkg.ErrInternal,
// 		},
// 		{
// 			name: "repo_get_exists",
// 			mock: func() {
// 				foundUser := &domain.User{Email: input.Email}
// 				s.userRepo.On("GetByEmail", mock.Anything, input.Email).Return(foundUser, nil).Once()
// 			},
// 			wantErr: pkg.ErrAlreadyExists,
// 		},
// 		{
// 			name: "repo_create_error",
// 			mock: func() {
// 				s.userRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, pkg.ErrNotFound)
// 				s.userRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()
// 			},
// 			wantErr: assert.AnError,
// 		},
// 		{
// 			name: "success",
// 			mock: func() {
// 				s.userRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, pkg.ErrNotFound)

// 				s.userRepo.On("Create",
// 					mock.Anything,
// 					mock.MatchedBy(func(u *domain.User) bool {
// 						return u.Email == input.Email &&
// 							u.Username == input.Username &&
// 							u.PasswordHash == input.PasswordHashed
// 					}),
// 				).Return(nil).Once()

// 				s.mediaSvc.On("GetPublicURL", mock.Anything).Return("http://mock-url.com").Twice()
// 			},
// 			wantErr: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		s.Run(tc.name, func() {
// 			s.userRepo.Calls = nil
// 			s.mediaSvc.Calls = nil
// 			s.userRepo.ExpectedCalls = nil
// 			s.mediaSvc.ExpectedCalls = nil

// 			tc.mock()

// 			got, err := s.userSvc.CreateUser(context.Background(), input)

// 			if tc.wantErr != nil {
// 				if tc.wantErr != assert.AnError {
// 					s.ErrorIs(err, tc.wantErr)
// 				} else {
// 					s.Error(err)
// 				}
// 			} else {
// 				s.NoError(err)
// 				s.Equal(input.Email, got.Email)
// 				s.Equal("http://mock-url.com", got.Avatar)
// 				s.Equal("http://mock-url.com", got.CoverImage)
// 			}

// 			s.userRepo.AssertExpectations(s.T())
// 			s.mediaSvc.AssertExpectations(s.T())
// 		})
// 	}
// }

// func (s *userServiceSuite) TestGetByEmail() {

// }

// // todo: implement more tests, and refactor auht_test, token_test
