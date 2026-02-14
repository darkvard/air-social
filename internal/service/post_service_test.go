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

type postServiceSuite struct {
	suite.Suite
}

func TestPostServiceSuite(t *testing.T) {
	suite.Run(t, new(postServiceSuite))
}

func (s *postServiceSuite) TestCreatePost() {
	var (
		userID      int64 = 1
		userSummary       = &domain.UserSummary{ID: userID}
		input             = domain.CreatePostParams{
			UserID:     userID,
			Content:    "Hello World",
			Visibility: domain.VisibilityPublic,
		}
		mediaInput = domain.CreatePostParams{
			UserID:     userID,
			Content:    "With Media",
			Visibility: domain.VisibilityPublic,
			Media: []domain.PostMediaParams{
				{MediaKey: "key1", MediaType: "image/jpeg"},
			},
		}
	)

	tests := []struct {
		name      string
		input     domain.CreatePostParams
		setupMock func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService)
		want      *domain.Post
		wantErr   error
	}{
		{
			name:  "user_not_found",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, input.UserID).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "invalid_data_empty",
			input: domain.CreatePostParams{
				UserID: userID,
			},
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInvalidData,
		},
		{
			name:  "media_verify_error",
			input: mediaInput,
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, mediaInput.UserID).Return(userSummary, nil).Once()
				mediaSvc.EXPECT().VerifyMedia(mock.Anything, []string{"key1"}).Return(pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "repo_create_error",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, input.UserID).Return(userSummary, nil).Once()
				postRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name:  "success_text_only",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, input.UserID).Return(userSummary, nil).Once()
				postRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *domain.Post) bool {
					return p.UserID == input.UserID && p.Content == input.Content
				})).Return(nil).Once()
			},
			want: &domain.Post{
				UserID:     input.UserID,
				Content:    input.Content,
				Visibility: input.Visibility,
				User:       userSummary,
			},
			wantErr: nil,
		},
		{
			name:  "success_with_media",
			input: mediaInput,
			setupMock: func(postRepo *mocks.PostRepository, mediaSvc *mocks.MediaService, userSvc *mocks.UserService) {
				userSvc.EXPECT().GetSummary(mock.Anything, mediaInput.UserID).Return(userSummary, nil).Once()
				mediaSvc.EXPECT().VerifyMedia(mock.Anything, []string{"key1"}).Return(nil).Once()
				postRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(p *domain.Post) bool {
					return len(p.Media) == 1 && p.Media[0].MediaKey == "key1"
				})).Return(nil).Once()
			},
			want: &domain.Post{
				UserID:     mediaInput.UserID,
				Content:    mediaInput.Content,
				Visibility: mediaInput.Visibility,
				User:       userSummary,
				Media: []domain.PostMedia{
					{MediaKey: "key1", MediaType: "image/jpeg"},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaSvc := mocks.NewMediaService(s.T())
			mockUserSvc := mocks.NewUserService(s.T())

			svc := NewPostService(mockPostRepo, mockMediaSvc, mockUserSvc)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockMediaSvc, mockUserSvc)
			}

			got, err := svc.CreatePost(context.Background(), tc.input)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.Content, got.Content)
				s.Equal(tc.want.UserID, got.UserID)
				if len(tc.want.Media) > 0 {
					s.Equal(tc.want.Media[0].MediaKey, got.Media[0].MediaKey)
				}
			}
		})
	}
}

func (s *postServiceSuite) TestGetPostDetail() {
	var (
		postID      int64 = 1
		userID      int64 = 10
		post              = &domain.Post{ID: postID, UserID: userID}
		userSummary       = &domain.UserSummary{ID: userID}
	)

	tests := []struct {
		name      string
		postID    int64
		setupMock func(postRepo *mocks.PostRepository, userSvc *mocks.UserService)
		want      *domain.Post
		wantErr   error
	}{
		{
			name:   "post_not_found",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userSvc *mocks.UserService) {
				postRepo.EXPECT().GetByID(mock.Anything, postID).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "user_service_error",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userSvc *mocks.UserService) {
				postRepo.EXPECT().GetByID(mock.Anything, postID).Return(post, nil).Once()
				userSvc.EXPECT().GetSummary(mock.Anything, userID).Return(nil, assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userSvc *mocks.UserService) {
				postRepo.EXPECT().GetByID(mock.Anything, postID).Return(post, nil).Once()
				userSvc.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
			},
			want: &domain.Post{ID: postID, UserID: userID, User: userSummary},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaSvc := mocks.NewMediaService(s.T())
			mockUserSvc := mocks.NewUserService(s.T())

			svc := NewPostService(mockPostRepo, mockMediaSvc, mockUserSvc)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockUserSvc)
			}

			got, err := svc.GetPostDetail(context.Background(), tc.postID)

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
