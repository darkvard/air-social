package comment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/comment"
	commentmocks "air-social/internal/domain/comment/mocks"
	"air-social/internal/domain/common"
	commonmocks "air-social/internal/domain/common/mocks"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type commentUseCaseSuite struct {
	suite.Suite
}

func TestCommentUseCaseSuite(t *testing.T) {
	suite.Run(t, new(commentUseCaseSuite))
}

func (s *commentUseCaseSuite) TestCreateComment() {
	var (
		postID   int64 = 100
		userID   int64 = 1
		authorID int64 = 2
		parentID int64 = 50
		rootID   int64 = 25
		content        = "Great post!"
	)

	// Common fixtures
	publicPost := &post.Post{
		ID:         postID,
		UserID:     authorID,
		Visibility: post.VisibilityPublic,
	}

	privatePost := &post.Post{
		ID:         postID,
		UserID:     authorID,
		Visibility: post.VisibilityPrivate,
	}

	followersPost := &post.Post{
		ID:         postID,
		UserID:     authorID,
		Visibility: post.VisibilityFollowers,
	}

	parentComment := &comment.Comment{
		ID:     parentID,
		PostID: postID,
	}

	subComment := &comment.Comment{
		ID:       parentID,
		PostID:   postID,
		ParentID: &rootID,
	}

	otherPostComment := &comment.Comment{
		ID:     parentID,
		PostID: 999,
	}

	mediaItems := []comment.Media{
		{MediaKey: "img1.jpg", MediaType: "image/jpeg"},
	}

	type testDeps struct {
		repo          *commentmocks.MockRepository
		postFetcher   *commentmocks.MockPostFetcher
		followChecker *commentmocks.MockFollowChecker
		mediaVerifier *commentmocks.MockMediaVerifier
		event         *commonmocks.MockEventPublisher
	}

	tests := []struct {
		name      string
		args      comment.CreateParams
		setupMock func(deps testDeps)
		want      *comment.Comment
		wantErr   error
	}{
		{
			name: "success_root_comment",
			args: comment.CreateParams{
				PostID:  postID,
				UserID:  userID,
				Content: content,
				Media:   mediaItems,
			},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"img1.jpg"}).
					Return(nil).Once()

				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(c *comment.Comment) bool {
						return c.PostID == postID && c.UserID == userID && c.ParentID == nil
					})).
					Return(nil).Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			want: &comment.Comment{
				PostID:  postID,
				UserID:  userID,
				Content: content,
				Media:   mediaItems,
			},
			wantErr: nil,
		},
		{
			name: "success_reply",
			args: comment.CreateParams{
				PostID:   postID,
				UserID:   userID,
				Content:  content,
				ParentID: &parentID,
			},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, parentID).
					Return(parentComment, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{}).
					Return(nil).Once()

				deps.repo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(c *comment.Comment) bool {
						return c.ParentID != nil && *c.ParentID == parentID
					})).
					Return(nil).Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			want: &comment.Comment{
				PostID:   postID,
				UserID:   userID,
				Content:  content,
				ParentID: &parentID,
			},
			wantErr: nil,
		},
		{
			name: "fail_reply_to_sub_comment",
			args: comment.CreateParams{
				PostID:   postID,
				UserID:   userID,
				Content:  content,
				ParentID: &parentID,
			},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, parentID).
					Return(subComment, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "post_fetch_error",
			args: comment.CreateParams{PostID: postID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "post_not_found_nil",
			args: comment.CreateParams{PostID: postID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(nil, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "private_post_forbidden",
			args: comment.CreateParams{PostID: postID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(privatePost, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "followers_post_not_following",
			args: comment.CreateParams{PostID: postID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(followersPost, nil).Once()

				deps.followChecker.EXPECT().
					GetRelationship(mock.Anything, userID, authorID).
					Return(follow.Relationship{IsFollowing: false}, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "parent_comment_wrong_post",
			args: comment.CreateParams{PostID: postID, UserID: userID, ParentID: &parentID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, parentID).
					Return(otherPostComment, nil).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "parent_comment_not_found",
			args: comment.CreateParams{PostID: postID, UserID: userID, ParentID: &parentID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.repo.EXPECT().
					GetByID(mock.Anything, parentID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "media_verification_failed",
			args: comment.CreateParams{PostID: postID, UserID: userID, Media: mediaItems},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"img1.jpg"}).
					Return(pkg.ErrNotFound).Once()
			},
			want:    nil,
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "repo_create_error",
			args: comment.CreateParams{PostID: postID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.postFetcher.EXPECT().
					GetPostDetail(mock.Anything, postID, userID).
					Return(publicPost, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{}).
					Return(nil).Once()

				deps.repo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(assert.AnError).Once()
			},
			want:    nil,
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := commentmocks.NewMockRepository(s.T())
			mockPostFetcher := commentmocks.NewMockPostFetcher(s.T())
			mockFollowChecker := commentmocks.NewMockFollowChecker(s.T())
			mockMediaVerifier := commentmocks.NewMockMediaVerifier(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:          mockRepo,
				postFetcher:   mockPostFetcher,
				followChecker: mockFollowChecker,
				mediaVerifier: mockMediaVerifier,
				event:         mockEvent,
			}

			uc := comment.NewUseCase(comment.Deps{
				CommentRepo:   mockRepo,
				PostFetcher:   mockPostFetcher,
				FollowChecker: mockFollowChecker,
				MediaVerifier: mockMediaVerifier,
				Event:         mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.CreateComment(context.Background(), tc.args)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.Equal(tc.want.PostID, got.PostID)
				s.Equal(tc.want.UserID, got.UserID)
				s.Equal(tc.want.Content, got.Content)
				if tc.want.ParentID != nil {
					s.NotNil(got.ParentID)
					s.Equal(*tc.want.ParentID, *got.ParentID)
				} else {
					s.Nil(got.ParentID)
				}
			}
		})
	}
}

func (s *commentUseCaseSuite) TestDeleteComment() {
	var (
		commentID int64 = 100
		userID    int64 = 1
		otherUser int64 = 2
	)

	ownComment := &comment.Comment{
		ID:     commentID,
		UserID: userID,
	}

	otherComment := &comment.Comment{
		ID:     commentID,
		UserID: otherUser,
	}

	type testDeps struct {
		repo  *commentmocks.MockRepository
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
			name: "success",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(ownComment, nil).Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, commentID).
					Return(nil).Once()

				deps.event.EXPECT().
					Publish(mock.Anything, mock.AnythingOfType("common.Event")).
					Return(nil).
					Once()
			},
			wantErr: nil,
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
					GetByID(mock.Anything, commentID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "repo_get_error",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(nil, assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "forbidden_not_owner",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(otherComment, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "repo_delete_error",
			args: struct {
				ctx       context.Context
				commentID int64
				userID    int64
			}{context.Background(), commentID, userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(ownComment, nil).Once()

				deps.repo.EXPECT().
					Delete(mock.Anything, commentID).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := commentmocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())

			deps := testDeps{
				repo:  mockRepo,
				event: mockEvent,
			}

			uc := comment.NewUseCase(comment.Deps{
				CommentRepo: mockRepo,
				Event:       mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			err := uc.DeleteComment(tc.args.ctx, tc.args.commentID, tc.args.userID)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *commentUseCaseSuite) TestUpdateComment() {
	var (
		commentID     int64 = 1
		userID        int64 = 10
		otherUserID   int64 = 99
		oldContent          = "Old content"
		newContent          = "New content"
		oldMedia            = []comment.Media{{MediaKey: "old.jpg", MediaType: "image"}}
		newMediaParam       = []comment.Media{{MediaKey: "new.jpg", MediaType: "image"}}
	)

	type testDeps struct {
		repo          *commentmocks.MockRepository
		mediaVerifier *commentmocks.MockMediaVerifier
	}

	tests := []struct {
		name      string
		args      comment.UpdateParams
		setupMock func(deps testDeps)
		want      *comment.Comment
		wantErr   error
	}{
		{
			name: "fail_comment_not_found",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(nil, pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "fail_forbidden",
			args: comment.UpdateParams{CommentID: commentID, UserID: otherUserID},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{ID: commentID, UserID: userID}, nil).Once()
			},
			wantErr: pkg.ErrForbidden,
		},
		{
			name: "fail_media_verification",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID, Media: newMediaParam},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{ID: commentID, UserID: userID}, nil).Once()
				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"new.jpg"}).
					Return(pkg.ErrNotFound).Once()
			},
			wantErr: pkg.ErrBadRequest,
		},
		{
			name: "fail_repo_update_error",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID, Content: &newContent},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{ID: commentID, UserID: userID}, nil).Once()
				deps.repo.EXPECT().
					Update(mock.Anything, mock.Anything).
					Return(assert.AnError).Once()
			},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_update_content_only",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID, Content: &newContent},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{
						ID:      commentID,
						UserID:  userID,
						Content: oldContent,
						Media:   oldMedia,
					}, nil).Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(c *comment.Comment) bool {
						return c.Content == newContent && len(c.Media) == 1 && c.Media[0].MediaKey == "old.jpg"
					})).
					Return(nil).Once()
			},
			want: &comment.Comment{
				Content: newContent,
				Media:   oldMedia,
			},
			wantErr: nil,
		},
		{
			name: "success_update_media_only",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID, Media: newMediaParam},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{
						ID:      commentID,
						UserID:  userID,
						Content: oldContent,
						Media:   oldMedia,
					}, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{"new.jpg"}).
					Return(nil).Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(c *comment.Comment) bool {
						return c.Content == oldContent && len(c.Media) == 1 && c.Media[0].MediaKey == "new.jpg"
					})).
					Return(nil).Once()
			},
			want: &comment.Comment{
				Content: oldContent,
				Media:   newMediaParam,
			},
			wantErr: nil,
		},
		{
			name: "success_clear_media",
			args: comment.UpdateParams{CommentID: commentID, UserID: userID, Media: []comment.Media{}},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetByID(mock.Anything, commentID).
					Return(&comment.Comment{
						ID:      commentID,
						UserID:  userID,
						Content: oldContent,
						Media:   oldMedia,
					}, nil).Once()

				deps.mediaVerifier.EXPECT().
					VerifyMedia(mock.Anything, []string{}).
					Return(nil).Once()

				deps.repo.EXPECT().
					Update(mock.Anything, mock.MatchedBy(func(c *comment.Comment) bool {
						return c.Content == oldContent && len(c.Media) == 0
					})).
					Return(nil).Once()
			},
			want: &comment.Comment{
				Content: oldContent,
				Media:   []comment.Media{},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := commentmocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockPostFetcher := commentmocks.NewMockPostFetcher(s.T())
			mockFollowChecker := commentmocks.NewMockFollowChecker(s.T())
			mockMediaVerifier := commentmocks.NewMockMediaVerifier(s.T())

			deps := testDeps{
				repo:          mockRepo,
				mediaVerifier: mockMediaVerifier,
			}

			uc := comment.NewUseCase(comment.Deps{
				CommentRepo:   mockRepo,
				PostFetcher:   mockPostFetcher,
				FollowChecker: mockFollowChecker,
				MediaVerifier: mockMediaVerifier,
				Event:         mockEvent,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.UpdateComment(context.Background(), tc.args)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Nil(got)
			} else {
				s.NoError(err)
				s.NotNil(got)
				s.Equal(tc.want.Content, got.Content)
				s.Equal(len(tc.want.Media), len(got.Media))
				if len(tc.want.Media) > 0 {
					s.Equal(tc.want.Media[0].MediaKey, got.Media[0].MediaKey)
				}
			}
		})
	}
}

func (s *commentUseCaseSuite) TestGetComments() {
	var (
		postID int64 = 100
		userID int64 = 1
	)

	params := comment.GetCursorParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Limit: 10,
		},
	}

	comments := []comment.Comment{
		{ID: 1, PostID: postID, Content: "c1"},
		{ID: 2, PostID: postID, Content: "c2"},
	}

	type testDeps struct {
		repo *commentmocks.MockRepository
		like *commentmocks.MockLikeChecker
	}

	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			postID int64
			params comment.GetCursorParams
		}
		setupMock func(deps testDeps)
		wantLen   int
		wantErr   error
	}{
		{
			name: "success",
			args: struct {
				ctx    context.Context
				postID int64
				params comment.GetCursorParams
			}{context.Background(), postID, params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetComments(mock.Anything, postID, mock.Anything).
					Return(comments, nil).Once()

				deps.like.EXPECT().
					IsCommentLiked(mock.Anything, []int64{1, 2}, userID).
					Return(map[int64]bool{1: true, 2: false}, nil).Once()
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "repo_error",
			args: struct {
				ctx    context.Context
				postID int64
				params comment.GetCursorParams
			}{context.Background(), postID, params},
			setupMock: func(deps testDeps) {
				deps.repo.EXPECT().
					GetComments(mock.Anything, postID, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantLen: 0,
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := commentmocks.NewMockRepository(s.T())
			mockEvent := commonmocks.NewMockEventPublisher(s.T())
			mockLike := commentmocks.NewMockLikeChecker(s.T())
			deps := testDeps{repo: mockRepo, like: mockLike}
			uc := comment.NewUseCase(comment.Deps{
				CommentRepo: mockRepo,
				Event:       mockEvent,
				LikeChecker: mockLike,
			})

			if tc.setupMock != nil {
				tc.setupMock(deps)
			}

			got, err := uc.GetComments(tc.args.ctx, tc.args.postID, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Len(got.Data, tc.wantLen)
				if tc.wantLen > 0 {
					// Verify IsLiked mapping from mock
					s.NotNil(got.Data[0].IsLiked)
					s.True(*got.Data[0].IsLiked)
				}
			}
		})
	}
}

func (s *commentUseCaseSuite) TestGetReplies() {
	var (
		parentID int64 = 50
		userID   int64 = 1
	)

	params := comment.GetCursorParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Limit: 10,
		},
	}

	replies := []comment.Comment{
		{ID: 51, ParentID: &parentID, Content: "r1"},
	}

	s.Run("success", func() {
		mockRepo := commentmocks.NewMockRepository(s.T())
		mockEvent := commonmocks.NewMockEventPublisher(s.T())
		mockLike := commentmocks.NewMockLikeChecker(s.T())
		uc := comment.NewUseCase(comment.Deps{CommentRepo: mockRepo, Event: mockEvent, LikeChecker: mockLike})

		mockRepo.EXPECT().
			GetReplies(mock.Anything, parentID, mock.Anything).
			Return(replies, nil).Once()

		mockLike.EXPECT().
			IsCommentLiked(mock.Anything, []int64{51}, userID).
			Return(map[int64]bool{51: false}, nil).Once()

		got, err := uc.GetReplies(context.Background(), parentID, params)
		s.NoError(err)
		s.Len(got.Data, 1)
	})

	s.Run("repo_error", func() {
		mockRepo := commentmocks.NewMockRepository(s.T())
		mockEvent := commonmocks.NewMockEventPublisher(s.T())
		uc := comment.NewUseCase(comment.Deps{CommentRepo: mockRepo, Event: mockEvent})

		mockRepo.EXPECT().
			GetReplies(mock.Anything, parentID, mock.Anything).
			Return(nil, assert.AnError).Once()

		_, err := uc.GetReplies(context.Background(), parentID, params)
		s.ErrorIs(err, pkg.ErrInternal)
	})
}
