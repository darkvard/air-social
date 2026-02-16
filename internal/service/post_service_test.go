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
		setupMock func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher)
		want      *domain.Post
		wantErr   error
	}{
		{
			name:  "user_not_found",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, input.UserID).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "invalid_data_empty",
			input: domain.CreatePostParams{
				UserID: userID,
			},
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInvalidData,
		},
		{
			name:  "media_verify_error",
			input: mediaInput,
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, mediaInput.UserID).Return(userSummary, nil).Once()
				mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"key1"}).Return(pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name:  "repo_create_error",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, input.UserID).Return(userSummary, nil).Once()
				postRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name:  "success_text_only",
			input: input,
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, input.UserID).Return(userSummary, nil).Once()
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
			setupMock: func(postRepo *mocks.PostRepository, mediaVerifier *mocks.MediaVerifier, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, mediaInput.UserID).Return(userSummary, nil).Once()
				mediaVerifier.EXPECT().VerifyMedia(mock.Anything, []string{"key1"}).Return(nil).Once()
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
			mockMediaVerifier := mocks.NewMediaVerifier(s.T())
			mockUserFetcher := mocks.NewUserSummaryFetcher(s.T())

			svc := NewPostService(mockPostRepo, mockMediaVerifier, mockUserFetcher)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockMediaVerifier, mockUserFetcher)
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

// todo
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
		setupMock func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher)
		want      *domain.Post
		wantErr   error
	}{
		{
			name:   "post_not_found",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().GetByID(mock.Anything, postID, -1).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "user_service_error",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().GetByID(mock.Anything, postID, -1).Return(post, nil).Once()
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(nil, assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			postID: postID,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().GetByID(mock.Anything, postID, -1).Return(post, nil).Once()
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
			},
			want: &domain.Post{ID: postID, UserID: userID, User: userSummary},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaVerifier := mocks.NewMediaVerifier(s.T())
			mockUserFetcher := mocks.NewUserSummaryFetcher(s.T())

			svc := NewPostService(mockPostRepo, mockMediaVerifier, mockUserFetcher)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockUserFetcher)
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

func (s *postServiceSuite) TestDeletePost() {
	var (
		postID int64 = 1
		userID int64 = 10
	)

	tests := []struct {
		name      string
		postID    int64
		userID    int64
		setupMock func(postRepo *mocks.PostRepository)
		wantErr   error
	}{
		{
			name:   "check_owner_error",
			postID: postID,
			userID: userID,
			setupMock: func(postRepo *mocks.PostRepository) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(false, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "post_not_found",
			postID: postID,
			userID: userID,
			setupMock: func(postRepo *mocks.PostRepository) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(false, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "not_owner",
			postID: postID,
			userID: userID,
			setupMock: func(postRepo *mocks.PostRepository) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(false, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "delete_error",
			postID: postID,
			userID: userID,
			setupMock: func(postRepo *mocks.PostRepository) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(true, nil).Once()
				postRepo.EXPECT().Delete(mock.Anything, postID).Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success",
			postID: postID,
			userID: userID,
			setupMock: func(postRepo *mocks.PostRepository) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(true, nil).Once()
				postRepo.EXPECT().Delete(mock.Anything, postID).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaVerifier := mocks.NewMediaVerifier(s.T())
			mockUserFetcher := mocks.NewUserSummaryFetcher(s.T())

			svc := NewPostService(mockPostRepo, mockMediaVerifier, mockUserFetcher)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo)
			}

			err := svc.DeletePost(context.Background(), tc.postID, tc.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *postServiceSuite) TestUpdatePost() {
	var (
		postID     int64 = 1
		userID     int64 = 10
		content          = "Updated content"
		visibility       = domain.VisibilityPrivate
		params           = domain.UpdatePostParams{
			PostID:     postID,
			UserID:     userID,
			Content:    &content,
			Visibility: &visibility,
		}
		existingPost = &domain.Post{
			ID:         postID,
			UserID:     userID,
			Content:    "Old content",
			Visibility: domain.VisibilityPublic,
		}
		userSummary = &domain.UserSummary{ID: userID}
	)

	tests := []struct {
		name      string
		params    domain.UpdatePostParams
		setupMock func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher)
		want      *domain.Post
		wantErr   error
	}{
		{
			name:   "check_owner_error",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(false, assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "not_owner",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(false, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrForbidden,
		},
		{
			name:   "get_post_error",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(true, nil).Once()
				postRepo.EXPECT().GetByID(mock.Anything, postID, userID).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "success",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				postRepo.EXPECT().IsOwner(mock.Anything, postID, userID).Return(true, nil).Once()
				postRepo.EXPECT().GetByID(mock.Anything, postID, userID).Return(existingPost, nil).Once()

				postRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *domain.Post) bool {
					return p.Content == content && p.Visibility == visibility
				})).Return(nil).Once()

				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
			},
			want: &domain.Post{
				ID:         postID,
				UserID:     userID,
				Content:    content,
				Visibility: visibility,
				User:       userSummary,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaVerifier := mocks.NewMediaVerifier(s.T())
			mockUserFetcher := mocks.NewUserSummaryFetcher(s.T())

			svc := NewPostService(mockPostRepo, mockMediaVerifier, mockUserFetcher)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockUserFetcher)
			}

			got, err := svc.UpdatePost(context.Background(), tc.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.Content, got.Content)
				s.Equal(tc.want.Visibility, got.Visibility)
				s.Equal(tc.want.User, got.User)
			}
		})
	}
}

func (s *postServiceSuite) TestGetUserPosts() {
	var (
		userID      int64 = 1
		limit       int   = 10
		cursor      int64 = 100
		params            = domain.CursorQueryParams{Limit: limit, Cursor: cursor}
		userSummary       = &domain.UserSummary{ID: userID}
	)

	tests := []struct {
		name      string
		params    domain.CursorQueryParams
		setupMock func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher)
		want      domain.CursorPaginatedResult[domain.Post]
		wantErr   error
	}{
		{
			name:   "user_not_found",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(nil, pkg.ErrNotFound).Once()
			},
			want:    domain.CursorPaginatedResult[domain.Post]{},
			wantErr: pkg.ErrNotFound,
		},
		{
			name:   "repo_error",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
				postRepo.EXPECT().GetUserPosts(mock.Anything, userID, params).Return(nil, assert.AnError).Once()
			},
			want:    domain.CursorPaginatedResult[domain.Post]{},
			wantErr: pkg.ErrInternal,
		},
		{
			name:   "success_has_next",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
				// Return limit + 1 items to simulate hasNextPage
				posts := make([]domain.Post, limit+1)
				for i := 0; i < limit+1; i++ {
					posts[i] = domain.Post{ID: int64(200 - i), UserID: userID}
				}
				postRepo.EXPECT().GetUserPosts(mock.Anything, userID, params).Return(posts, nil).Once()
			},
			want: domain.CursorPaginatedResult[domain.Post]{
				NextCursor:  int64(200 - (limit - 1)), // ID of the last item in the sliced list
				HasNextPage: true,
			},
			wantErr: nil,
		},
		{
			name:   "success_no_next",
			params: params,
			setupMock: func(postRepo *mocks.PostRepository, userFetcher *mocks.UserSummaryFetcher) {
				userFetcher.EXPECT().GetSummary(mock.Anything, userID).Return(userSummary, nil).Once()
				// Return exactly limit items
				posts := make([]domain.Post, limit)
				for i := 0; i < limit; i++ {
					posts[i] = domain.Post{ID: int64(200 - i), UserID: userID}
				}
				postRepo.EXPECT().GetUserPosts(mock.Anything, userID, params).Return(posts, nil).Once()
			},
			want: domain.CursorPaginatedResult[domain.Post]{
				NextCursor:  0,
				HasNextPage: false,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockPostRepo := mocks.NewPostRepository(s.T())
			mockMediaVerifier := mocks.NewMediaVerifier(s.T())
			mockUserFetcher := mocks.NewUserSummaryFetcher(s.T())

			svc := NewPostService(mockPostRepo, mockMediaVerifier, mockUserFetcher)

			if tc.setupMock != nil {
				tc.setupMock(mockPostRepo, mockUserFetcher)
			}

			got, err := svc.GetUserPosts(context.Background(), userID, tc.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.Data)
			} else {
				s.NoError(err)
				s.Equal(tc.want.HasNextPage, got.HasNextPage)
				s.Equal(tc.want.NextCursor, got.NextCursor)
				if tc.want.HasNextPage {
					s.Len(got.Data, limit)
				} else {
					s.Len(got.Data, limit) // In success_no_next case we mocked exactly limit items
				}
				// Verify user enrichment
				for _, p := range got.Data {
					s.Equal(userSummary, p.User)
				}
			}
		})
	}
}
