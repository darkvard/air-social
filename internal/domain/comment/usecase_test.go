package comment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/comment"
	commentmocks "air-social/internal/domain/comment/mocks"
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

			deps := testDeps{
				repo:          mockRepo,
				postFetcher:   mockPostFetcher,
				followChecker: mockFollowChecker,
				mediaVerifier: mockMediaVerifier,
			}

			uc := comment.NewUseCase(comment.Deps{
				CommentRepo:   mockRepo,
				PostFetcher:   mockPostFetcher,
				FollowChecker: mockFollowChecker,
				MediaVerifier: mockMediaVerifier,
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
