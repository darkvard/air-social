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

func (s *LikeUseCaseSuite) TestIsPostLiked() {
	var (
		userID  int64 = 1
		postIDs       = []int64{10, 20, 30}
	)

	type testDeps struct {
		repo *likemocks.MockRepository
	}

	tests := []struct {
		name      string
		args      struct {
			ctx     context.Context
			postIDs []int64
			userID  int64
		}
		setupMock func(deps testDeps)
		want      map[int64]bool
		wantErr   error
	}{
		{
			name: "success",
			args: struct {
				ctx     context.Context
				postIDs []int64
				userID  int64
			}{context.Background(), postIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostLiked(mock.Anything, postIDs, userID).
					Return([]int64{10, 30}, nil).
					Once()
			},
			want: map[int64]bool{
				10: true,
				20: false,
				30: true,
			},
			wantErr: nil,
		},
		{
			name: "repo_error",
			args: struct {
				ctx     context.Context
				postIDs []int64
				userID  int64
			}{context.Background(), postIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostLiked(mock.Anything, postIDs, userID).
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_no_liked_posts",
			args: struct {
				ctx     context.Context
				postIDs []int64
				userID  int64
			}{context.Background(), postIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetPostLiked(mock.Anything, postIDs, userID).
					Return([]int64{}, nil).
					Once()
			},
			want: map[int64]bool{
				10: false,
				20: false,
				30: false,
			},
			wantErr: nil,
		},
		{
			name: "success_empty_input_ids",
			args: struct {
				ctx     context.Context
				postIDs []int64
				userID  int64
			}{context.Background(), []int64{}, userID},
			setupMock: func(deps testDeps) {
			},
			want:    map[int64]bool{},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.IsPostLiked(tc.args.ctx, tc.args.postIDs, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}

func (s *LikeUseCaseSuite) TestIsCommentLiked() {
	var (
		userID     int64 = 1
		commentIDs       = []int64{101, 102, 103}
	)

	type testDeps struct {
		repo *likemocks.MockRepository
	}

	tests := []struct {
		name      string
		args      struct {
			ctx        context.Context
			commentIDs []int64
			userID     int64
		}
		setupMock func(deps testDeps)
		want      map[int64]bool
		wantErr   error
	}{
		{
			name: "success",
			args: struct {
				ctx        context.Context
				commentIDs []int64
				userID     int64
			}{context.Background(), commentIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetCommentLiked(mock.Anything, commentIDs, userID).
					Return([]int64{102}, nil).
					Once()
			},
			want: map[int64]bool{
				101: false,
				102: true,
				103: false,
			},
			wantErr: nil,
		},
		{
			name: "repo_error",
			args: struct {
				ctx        context.Context
				commentIDs []int64
				userID     int64
			}{context.Background(), commentIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetCommentLiked(mock.Anything, commentIDs, userID).
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_no_liked_comments",
			args: struct {
				ctx        context.Context
				commentIDs []int64
				userID     int64
			}{context.Background(), commentIDs, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetCommentLiked(mock.Anything, commentIDs, userID).
					Return([]int64{}, nil).
					Once()
			},
			want: map[int64]bool{
				101: false,
				102: false,
				103: false,
			},
			wantErr: nil,
		},
		{
			name: "success_empty_input_ids",
			args: struct {
				ctx        context.Context
				commentIDs []int64
				userID     int64
			}{context.Background(), []int64{}, userID},
			setupMock: func(deps testDeps) {
			},
			want:    map[int64]bool{},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := likemocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := like.NewUsecase(like.Deps{
				Repo:  mockRepo,
				Event: mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.IsCommentLiked(tc.args.ctx, tc.args.commentIDs, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want, got)
			}
		})
	}
}
