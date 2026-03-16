package post_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/post"
	postmocks "air-social/internal/domain/post/mocks"
	"air-social/internal/domain/stats"
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
		repo         *postmocks.MockRepository
		statsFetcher *postmocks.MockStatsFetcher
		likeChecker  *postmocks.MockLikeChecker
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
			name: "success_with_metadata",
			args: args{ctx: context.Background(), postID: postID, viewerID: viewerID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetDetail(mock.Anything, postID).
					Return(expectedPost, nil).
					Once()

				deps.statsFetcher.EXPECT().
					GetPostsStats(mock.Anything, []int64{postID}).
					Return(map[int64]stats.PostStats{
						postID: {PostID: postID, LikesCount: 10, CommentsCount: 5, SharesCount: 2},
					}, nil).
					Once()

				deps.likeChecker.EXPECT().
					IsPostLiked(mock.Anything, []int64{postID}, viewerID).
					Return(map[int64]bool{postID: true}, nil).
					Once()
			},
			want:    expectedPost,
			wantErr: nil,
		},
		{
			name: "success_metadata_error_ignored",
			args: args{ctx: context.Background(), postID: postID, viewerID: viewerID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetDetail(mock.Anything, postID).
					Return(expectedPost, nil).
					Once()

				deps.statsFetcher.EXPECT().
					GetPostsStats(mock.Anything, []int64{postID}).
					Return(nil, assert.AnError).
					Once()

				deps.likeChecker.EXPECT().
					IsPostLiked(mock.Anything, []int64{postID}, viewerID).
					Return(nil, assert.AnError).
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
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{
				repo:         mockRepo,
				statsFetcher: mockStats,
				likeChecker:  mockLike,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
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
				s.Equal(tc.want.ID, got.ID)
				if got.Stat.LikesCount > 0 {
					s.Equal(int32(10), got.Stat.LikesCount)
					s.Equal(int32(5), got.Stat.CommentsCount)
					s.NotNil(got.IsLiked)
					s.True(*got.IsLiked)
				}
			}
		})
	}
}

func (s *postUseCaseSuite) TestCreatePost() {
	var (
		userID         = int64(1)
		content        = "Hello World"
		originalPostID = int64(999)
	)

	mediaParams := []post.MediaParams{
		{MediaKey: "key1", MediaType: "image/png"},
	}

	type testDeps struct {
		repo     *postmocks.MockRepository
		verifier *postmocks.MockMediaVerifier
		event    *commonmocks.MockEventPublisher
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
			wantErr: pkg.ErrBadRequest,
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
		{
			name: "success_share_post",
			args: args{
				ctx: context.Background(),
				params: post.CreateParams{
					UserID:         userID,
					Content:        content,
					OriginalPostID: &originalPostID,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.UserID == userID && p.OriginalPostID != nil && *p.OriginalPostID == originalPostID
					})).
					Return(nil).
					Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			want: &post.Post{
				UserID:         userID,
				Content:        content,
				OriginalPostID: &originalPostID,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{
				repo:     mockRepo,
				verifier: mockVerifier,
				event:    mockEvent,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
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
		postID        = int64(1)
		userID        = int64(10)
		otherUserID   = int64(99)
		newContent    = "Updated Content"
		newVisibility = post.VisibilityFollowers
	)

	mediaParams := []post.MediaParams{
		{MediaKey: "new-key", MediaType: "image/jpeg"},
	}

	type testDeps struct {
		repo     *postmocks.MockRepository
		verifier *postmocks.MockMediaVerifier
		event    *commonmocks.MockEventPublisher
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
					Return(&post.Post{ID: postID, UserID: userID}, nil).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "media_verification_failed",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID: postID,
					UserID: userID,
					Media:  mediaParams,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(&post.Post{ID: postID, UserID: userID}, nil).
					Once()

				deps.verifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"new-key"}).
					Return(pkg.ErrNotFound).
					Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
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
					Return(&post.Post{ID: postID, UserID: userID, Content: "Old"}, nil).
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
			name: "success_content_only",
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
					Return(&post.Post{ID: postID, UserID: userID, Content: "Old", Visibility: post.VisibilityPublic}, nil).
					Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.ID == postID && p.Content == newContent && p.Media == nil
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				ID:         postID,
				UserID:     userID,
				Content:    newContent,
				Visibility: post.VisibilityPublic, // Ensure visibility is preserved
			},
			wantErr: nil,
		},
		{
			name: "success_with_media_and_visibility",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID:     postID,
					UserID:     userID,
					Visibility: &newVisibility,
					Media:      mediaParams,
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(&post.Post{ID: postID, UserID: userID, Content: "Old Content", Visibility: post.VisibilityPublic}, nil).
					Once()

				deps.verifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"new-key"}).
					Return(nil).
					Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.ID == postID &&
							p.Visibility == newVisibility &&
							len(p.Media) == 1 &&
							p.Media[0].MediaKey == "new-key"
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				ID:         postID,
				UserID:     userID,
				Content:    "Old Content", // Ensure content is preserved
				Visibility: newVisibility,
				Media: []post.Media{
					{MediaKey: "new-key", MediaType: "image/jpeg"},
				},
			},
			wantErr: nil,
		},
		{
			name: "success_clear_media",
			args: args{
				ctx: context.Background(),
				params: post.UpdateParams{
					PostID: postID,
					UserID: userID,
					Media:  []post.MediaParams{},
				},
			},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(&post.Post{
						ID:     postID,
						UserID: userID,
						Media:  []post.Media{{MediaKey: "old"}},
					}, nil).
					Once()

				deps.verifier.EXPECT().
					VerifyMedia(mock.Anything, []string{}).
					Return(nil).
					Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(p *post.Post) bool {
						return p.ID == postID && p.Media != nil && len(p.Media) == 0
					})).
					Return(nil).
					Once()
			},
			want: &post.Post{
				ID:     postID,
				UserID: userID,
				Media:  []post.Media{},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{
				repo:     mockRepo,
				verifier: mockVerifier,
				event:    mockEvent,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
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
				if tc.want.Content != "" {
					s.Equal(tc.want.Content, got.Content)
				}
				if tc.want.Visibility != "" {
					s.Equal(tc.want.Visibility, got.Visibility)
				}
				if tc.want.Media != nil {
					s.Equal(len(tc.want.Media), len(got.Media))
					if len(tc.want.Media) > 0 {
						s.Equal(tc.want.Media[0].MediaKey, got.Media[0].MediaKey)
						s.Equal(tc.want.Media[0].MediaType, got.Media[0].MediaType)
					}
				}
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

	originalPostID := int64(888)

	existingPost := &post.Post{
		ID:     postID,
		UserID: userID,
	}
	sharedPost := &post.Post{
		ID:             postID,
		UserID:         userID,
		OriginalPostID: &originalPostID,
	}

	type testDeps struct {
		repo  *postmocks.MockRepository
		event *commonmocks.MockEventPublisher
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
		{
			name: "success_shared_post",
			args: args{ctx: context.Background(), postID: postID, userID: userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, postID).
					Return(sharedPost, nil).
					Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, postID).
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
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
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
		repo         *postmocks.MockRepository
		statsFetcher *postmocks.MockStatsFetcher
		likeChecker  *postmocks.MockLikeChecker
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
			name: "success_with_metadata",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUserPosts(mock.Anything, mock.Anything).
					Return(posts, nil).
					Once()

				deps.statsFetcher.EXPECT().
					GetPostsStats(mock.Anything, []int64{1, 2}).
					Return(map[int64]stats.PostStats{
						1: {PostID: 1, LikesCount: 15},
						2: {PostID: 2, CommentsCount: 3},
					}, nil).
					Once()

				deps.likeChecker.EXPECT().
					IsPostLiked(mock.Anything, []int64{1, 2}, userID).
					Return(map[int64]bool{1: true, 2: false}, nil).
					Once()
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "success_metadata_error_ignored",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUserPosts(mock.Anything, mock.Anything).
					Return(posts, nil).
					Once()

				deps.statsFetcher.EXPECT().
					GetPostsStats(mock.Anything, []int64{1, 2}).
					Return(nil, assert.AnError).
					Once()

				deps.likeChecker.EXPECT().
					IsPostLiked(mock.Anything, []int64{1, 2}, userID).
					Return(nil, assert.AnError).
					Once()
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "success_empty_result",
			args: args{ctx: context.Background(), params: params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetUserPosts(mock.Anything, mock.Anything).
					Return([]post.Post{}, nil).
					Once()
			},
			wantLen: 0,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := postmocks.NewMockRepository(s.T())
			mockVerifier := postmocks.NewMockMediaVerifier(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{
				repo:         mockRepo,
				statsFetcher: mockStats,
				likeChecker:  mockLike,
			}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
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
				if tc.wantLen > 0 && got.Data[0].Stat.LikesCount > 0 {
					s.Equal(int32(15), got.Data[0].Stat.LikesCount)
					s.NotNil(got.Data[0].IsLiked)
					s.True(*got.Data[0].IsLiked)
					s.Equal(int32(3), got.Data[1].Stat.CommentsCount)
					s.NotNil(got.Data[1].IsLiked)
					s.False(*got.Data[1].IsLiked)
				}
			}
		})
	}
}

func (s *postUseCaseSuite) TestGetPostSharers() {
	var (
		postID = int64(1)
		limit  = 10
	)

	params := post.GetSharersParams{
		PostID: postID,
		Query: common.CursorQueryParams[int64]{
			Limit: limit,
		},
	}

	sharers := []post.Sharer{
		{ShareID: 101, UserID: 2, FullName: "User Two"},
		{ShareID: 102, UserID: 3, FullName: "User Three"},
	}

	type testDeps struct {
		repo *postmocks.MockRepository
	}

	type args struct {
		ctx    context.Context
		params post.GetSharersParams
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
					GetPostSharers(mock.Anything, params).
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
					GetPostSharers(mock.Anything, params).
					Return(sharers, nil).
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
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockStats := postmocks.NewMockStatsFetcher(s.T())
			mockLike := postmocks.NewMockLikeChecker(s.T())

			deps := testDeps{repo: mockRepo}

			uc := post.NewUseCase(post.Deps{
				PostRepo:      mockRepo,
				MediaVerifier: mockVerifier,
				Event:         mockEvent,
				StatsFetcher:  mockStats,
				LikeChecker:   mockLike,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetPostSharers(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.Data)
			} else {
				s.NoError(err)
				s.Len(got.Data, tc.wantLen)
				if tc.wantLen > 0 {
					s.Equal(sharers[0].UserID, got.Data[0].UserID)
					s.Equal(sharers[0].ShareID, got.Data[0].GetCursor())
				}
			}
		})
	}
}
