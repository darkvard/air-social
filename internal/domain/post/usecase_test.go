package post_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	postmocks "air-social/internal/domain/post/mocks"
	"air-social/pkg"
)

type postUseCaseSuite struct {
	suite.Suite
}

func TestPostUseCaseSuite(t *testing.T) {
	suite.Run(t, new(postUseCaseSuite))
}

func (s *postUseCaseSuite) TestGetPostDetail() {
	var (
		postID   = int64(1)
		viewerID = int64(2)
	)

	expectedPost := &post.Post{
		ID:     postID,
		UserID: int64(3),
	}

	type testDeps struct {
		repo *postmocks.MockRepository
	}

	type args struct {
		ctx      context.Context
		postID   int64
		viewerID int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      *post.Post
		wantErr   error
	}{
		{
			name: "repo_error_not_found",
			args: args{ctx: context.Background(), postID: postID, viewerID: viewerID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetDetail(mock.Anything, postID).
					Return(nil, pkg.ErrNotFound).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "repo_error_internal",
			args: args{ctx: context.Background(), postID: postID, viewerID: viewerID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetDetail(mock.Anything, postID).
					Return(nil, assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), postID: postID, viewerID: viewerID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetDetail(mock.Anything, postID).
					Return(expectedPost, nil).
					Once()
			},
			want:    expectedPost,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetPostDetail(tc.args.ctx, tc.args.postID, tc.args.viewerID)

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

func (s *postUseCaseSuite) TestCreatePost() {
	var (
		userID  = int64(1)
		content = "Hello World"
	)

	mediaParams := []post.MediaParams{
		{MediaKey: "key1", MediaType: "image/png"},
	}

	type testDeps struct {
		repo     *postmocks.MockRepository
		verifier *postmocks.MockMediaVerifier
	}

	type args struct {
		ctx    context.Context
		params post.CreateParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      *post.Post
		wantErr   error
	}{
		{
			name: "invalid_data_empty",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:  userID,
					Content: "   ",
					Media:   nil,
				},
			},
			setupMock: func(deps testDeps) {},
			want:      nil,
			wantErr:   pkg.ErrInvalidData,
		},
		{
			name: "media_verification_failed",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:  userID,
					Content: content,
					Media:   mediaParams,
				},
			},
			setupMock: func(deps testDeps) {
				deps.verifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"key1"}).
					Return(pkg.ErrNotFound).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "repo_create_error",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:  userID,
					Content: content,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.UserID == userID && p.Content == content
					})).
					Return(assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: assert.AnError,
		},
		{
			name: "success_text_only",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:     userID,
					Content:    content,
					Visibility: post.VisibilityPublic,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.UserID == userID && p.Content == content && p.Visibility == post.VisibilityPublic
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				UserID:     userID,
				Content:    content,
				Visibility: post.VisibilityPublic,
			},
			wantErr: nil,
		},
		{
			name: "success_with_media",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:  userID,
					Content: content,
					Media:   mediaParams,
				},
			},
			setupMock: func(deps testDeps) {
				deps.verifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"key1"}).
					Return(nil).
					Once()

				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.UserID == userID && len(p.Media) == 1 && p.Media[0].MediaKey == "key1"
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				UserID:  userID,
				Content: content,
				Media: []post.Media{
					{MediaKey: "key1", MediaType: "image/png"},
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo:     mockRepo,
				verifier: mockVerifier,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.CreatePost(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.UserID, got.UserID)
				s.Equal(tc.want.Content, got.Content)
				if len(tc.want.Media) > 0 {
					s.Len(got.Media, len(tc.want.Media))
					s.Equal(tc.want.Media[0].MediaKey, got.Media[0].MediaKey)
				}
			}
		})
	}
}

func (s *postUseCaseSuite) TestUpdatePost() {
	var (
		postID      = int64(1)
		userID      = int64(10)
		otherUserID = int64(99)
		newContent  = "Updated Content"
	)

	existingPost := &post.Post{
		ID:      postID,
		UserID:  userID,
		Content: "Old Content",
	}

	type testDeps struct {
		repo *postmocks.MockRepository
	}

	type args struct {
		ctx    context.Context
		params post.UpdateParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		want      *post.Post
		wantErr   error
	}{
		{
			name: "post_not_found",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID: postID,
					UserID: userID,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(nil, pkg.ErrNotFound).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "forbidden",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID: postID,
					UserID: otherUserID,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "repo_update_error",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID:  postID,
					UserID:  userID,
					Content: &newContent,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.ID == postID && p.Content == newContent
					})).
					Return(assert.AnError).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID:  postID,
					UserID:  userID,
					Content: &newContent,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.ID == postID && p.Content == newContent
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				ID:      postID,
				UserID:  userID,
				Content: newContent,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdatePost(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.Content, got.Content)
			}
		})
	}
}

func (s *postUseCaseSuite) TestDeletePost() {
	var (
		postID      = int64(1)
		userID      = int64(10)
		otherUserID = int64(99)
	)

	existingPost := &post.Post{
		ID:     postID,
		UserID: userID,
	}

	type testDeps struct {
		repo *postmocks.MockRepository
	}

	type args struct {
		ctx    context.Context
		postID int64
		userID int64
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantErr   error
	}{
		{
			name: "post_not_found",
			args: args{ctx: context.Background(), postID: postID, userID: userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(nil, pkg.ErrNotFound).
					Once()
			},
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "forbidden",
			args: args{ctx: context.Background(), postID: postID, userID: otherUserID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "repo_delete_error",
			args: args{ctx: context.Background(), postID: postID, userID: userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, postID).
					Return(assert.AnError).
					Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), postID: postID, userID: userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(existingPost, nil).
					Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, postID).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.DeletePost(tc.args.ctx, tc.args.postID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *postUseCaseSuite) TestGetUserPosts() {
	var (
		userID = int64(1)
		limit  = 10
	)

	params := post.GetCursorParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Limit: limit,
		},
	}

	posts := []post.Post{
		{ID: 1, UserID: userID},
		{ID: 2, UserID: userID},
	}

	type testDeps struct {
		repo *postmocks.MockRepository
	}

	type args struct {
		ctx    context.Context
		params post.GetCursorParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name: "repo_error",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUserPosts(mock.Anything, mock.Anything).
					Return(nil, pkg.ErrNotFound).
					Once()
			},
			wantLen: 0,
			wantErr: pkg.ErrNotFound,
		},
		{
			name: "success",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUserPosts(mock.Anything, mock.Anything).
					Return(posts, nil).
					Once()
			},
			wantLen: 2,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo: mockRepo,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetUserPosts(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.Data)
			} else {
				s.NoError(err)
				s.Len(got.Data, tc.wantLen)
			}
		})
	}
}
