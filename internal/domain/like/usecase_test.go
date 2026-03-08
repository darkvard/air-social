package like_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/like"
	likemocks "air-social/internal/domain/like/mocks"
	"air-social/pkg"
)

type LikeUseCaseSuite struct {
	suite.Suite
}

func TestLikeUseCaseSuite(t *testing.T) {
	suite.Run(t, new(LikeUseCaseSuite))
}

func (s *LikeUseCaseSuite) TestLikePost() {
	var (
		postID int64 = 100
		userID int64 = 1
	)

	type testDeps struct {
		repo  *likemocks.MockRepository
		event *commonmocks.MockEventPublisher
	}

	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			postID int64
			userID int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "repo_internal_error",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertPostLike(mock.Anything, postID, userID).
					Return(false, int64(0), assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "post_not_found",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertPostLike(mock.Anything, postID, userID).
					Return(false, int64(0), pkg.ErrInvalidData).
					Once()
			},
			wantErr: pkg.ErrInvalidData,
		},
		{
			name: "success_new_like",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertPostLike(mock.Anything, postID, userID).
					Return(true, int64(999), nil).
					Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "success_already_liked",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertPostLike(mock.Anything, postID, userID).
					Return(false, int64(0), nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.LikePost(tc.args.ctx, tc.args.postID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *LikeUseCaseSuite) TestUnlikePost() {
	var (
		postID int64 = 100
		userID int64 = 1
	)

	type testDeps struct {
		repo  *likemocks.MockRepository
		event *commonmocks.MockEventPublisher
	}

	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			postID int64
			userID int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "repo_error",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					DeletePostLike(mock.Anything, postID, userID).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: struct {
				ctx    context.Context
				postID int64
				userID int64
			}{context.Background(), postID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					DeletePostLike(mock.Anything, postID, userID).
					Return(nil).
					Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.UnlikePost(tc.args.ctx, tc.args.postID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *LikeUseCaseSuite) TestLikeComment() {
	var (
		commentID int64 = 200
		userID    int64 = 1
	)

	type testDeps struct {
		repo  *likemocks.MockRepository
		event *commonmocks.MockEventPublisher
	}

	tests := []struct {
		name string
		args struct {
			ctx       context.Context
			commentID int64
			userID    int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "repo_internal_error",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertCommentLike(mock.Anything, commentID, userID).
					Return(false, int64(0), assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "comment_not_found",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertCommentLike(mock.Anything, commentID, userID).
					Return(false, int64(0), pkg.ErrInvalidData).
					Once()
			},
			wantErr: pkg.ErrInvalidData,
		},
		{
			name: "success_new_like",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertCommentLike(mock.Anything, commentID, userID).
					Return(true, int64(999), nil).
					Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
		{
			name: "success_already_liked",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					InsertCommentLike(mock.Anything, commentID, userID).
					Return(false, int64(0), nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.LikeComment(tc.args.ctx, tc.args.commentID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *LikeUseCaseSuite) TestUnlikeComment() {
	var (
		commentID int64 = 200
		userID    int64 = 1
	)

	type testDeps struct {
		repo  *likemocks.MockRepository
		event *commonmocks.MockEventPublisher
	}

	tests := []struct {
		name string
		args struct {
			ctx       context.Context
			commentID int64
			userID    int64
		}
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "repo_error",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					DeleteCommentLike(mock.Anything, commentID, userID).
					Return(assert.AnError).
					Once()
			},
			wantErr: assert.AnError,
		},
		{
			name: "success",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					DeleteCommentLike(mock.Anything, commentID, userID).
					Return(nil).
					Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.UnlikeComment(tc.args.ctx, tc.args.commentID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}
