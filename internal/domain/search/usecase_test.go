package search_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/common"
	"air-social/internal/domain/search"
	searchmocks "air-social/internal/domain/search/mocks"
	"air-social/pkg"
)

type searchUseCaseSuite struct {
	suite.Suite
}

func TestSearchUseCaseSuite(t *testing.T) {
	suite.Run(t, new(searchUseCaseSuite))
}

// --- SearchUsers ---

func (s *searchUseCaseSuite) TestSearchUsers() {
	type args struct {
		ctx    context.Context
		params search.UsersParams
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(repo *searchmocks.MockRepository)
		want      common.CursorPaginatedResult[search.User, int64]
		wantErr   error
	}{
		{
			name: "repo_error_returns_internal",
			args: args{
				ctx: context.Background(),
				params: search.UsersParams{
					Search: "alice",
					Query:  common.CursorQueryParams[int64]{Limit: 10},
				},
			},
			setupMock: func(repo *searchmocks.MockRepository) {
				repo.EXPECT().
					SearchUsers(mock.Anything, mock.Anything).
					Return(nil, assert.AnError).
					Once()
			},
			want:    common.CursorPaginatedResult[search.User, int64]{},
			wantErr: pkg.ErrInternal,
		},
		{
			name: "success_empty_result",
			args: args{
				ctx: context.Background(),
				params: search.UsersParams{
					Search: "nobody",
					Query:  common.CursorQueryParams[int64]{Limit: 5},
				},
			},
			setupMock: func(repo *searchmocks.MockRepository) {
				repo.EXPECT().
					SearchUsers(mock.Anything, mock.Anything).
					Return([]search.User{}, nil).
					Once()
			},
			want: common.CursorPaginatedResult[search.User, int64]{
				Data:        []search.User{},
				HasNextPage: false,
				NextCursor:  0,
			},
			wantErr: nil,
		},
		{
			name: "success_no_next_page",
			args: args{
				ctx: context.Background(),
				params: search.UsersParams{
					Search: "alice",
					Query:  common.CursorQueryParams[int64]{Limit: 3},
				},
			},
			setupMock: func(repo *searchmocks.MockRepository) {
				// repo trả về <= limit records → không có trang tiếp
				repo.EXPECT().
					SearchUsers(mock.Anything, mock.Anything).
					Return([]search.User{
						{ID: 1, Username: "alice1"},
						{ID: 2, Username: "alice2"},
					}, nil).
					Once()
			},
			want: common.CursorPaginatedResult[search.User, int64]{
				Data: []search.User{
					{ID: 1, Username: "alice1"},
					{ID: 2, Username: "alice2"},
				},
				HasNextPage: false,
				NextCursor:  0,
			},
			wantErr: nil,
		},
		{
			name: "success_has_next_page",
			args: args{
				ctx: context.Background(),
				params: search.UsersParams{
					Search: "alice",
					Query:  common.CursorQueryParams[int64]{Limit: 2},
				},
			},
			setupMock: func(repo *searchmocks.MockRepository) {
				// repo trả về limit+1 records → có trang tiếp
				repo.EXPECT().
					SearchUsers(mock.Anything, mock.Anything).
					Return([]search.User{
						{ID: 10, Username: "alice1"},
						{ID: 20, Username: "alice2"},
						{ID: 30, Username: "alice3"}, // record dư để detect next page
					}, nil).
					Once()
			},
			want: common.CursorPaginatedResult[search.User, int64]{
				Data: []search.User{
					{ID: 10, Username: "alice1"},
					{ID: 20, Username: "alice2"},
				},
				HasNextPage: true,
				NextCursor:  20,
			},
			wantErr: nil,
		},
		{
			name: "invalid_limit_normalized_to_default",
			args: args{
				ctx: context.Background(),
				params: search.UsersParams{
					Search: "alice",
					Query:  common.CursorQueryParams[int64]{Limit: 0}, // dưới minimum, sẽ normalize thành 10
				},
			},
			setupMock: func(repo *searchmocks.MockRepository) {
				repo.EXPECT().
					SearchUsers(mock.Anything, mock.Anything).
					Return([]search.User{{ID: 1, Username: "alice1"}}, nil).
					Once()
			},
			want: common.CursorPaginatedResult[search.User, int64]{
				Data:        []search.User{{ID: 1, Username: "alice1"}},
				HasNextPage: false,
				NextCursor:  0,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := searchmocks.NewMockRepository(s.T())
			if tc.setupMock != nil {
				tc.setupMock(mockRepo)
			}

			uc := search.NewUseCase(mockRepo)
			got, err := uc.SearchUsers(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
				s.Empty(got.Data)
			} else {
				s.NoError(err)
				s.Equal(tc.want.HasNextPage, got.HasNextPage)
				s.Equal(tc.want.NextCursor, got.NextCursor)
				s.Equal(len(tc.want.Data), len(got.Data))
				for i := range tc.want.Data {
					s.Equal(tc.want.Data[i].ID, got.Data[i].ID)
				}
			}
		})
	}
}

// --- SearchPosts ---

func (s *searchUseCaseSuite) TestSearchPosts() {
	type args struct {
		ctx    context.Context
		params search.PostsParams
	}

	tests := []struct {
		name    string
		args    args
		want    common.CursorPaginatedResult[search.Post, int64]
		wantErr error
	}{
		{
			name: "stub_returns_empty",
			args: args{
				ctx: context.Background(),
				params: search.PostsParams{
					Search: "hello",
					Query:  common.CursorQueryParams[int64]{Limit: 10},
				},
			},
			want: common.CursorPaginatedResult[search.Post, int64]{
				Data:        nil,
				HasNextPage: false,
				NextCursor:  0,
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockRepo := searchmocks.NewMockRepository(s.T())

			uc := search.NewUseCase(mockRepo)
			got, err := uc.SearchPosts(tc.args.ctx, tc.args.params)

			if tc.wantErr != nil {
				s.ErrorIs(err, tc.wantErr)
			} else {
				s.NoError(err)
				s.Equal(tc.want.HasNextPage, got.HasNextPage)
				s.Equal(tc.want.NextCursor, got.NextCursor)
			}
		})
	}
}
