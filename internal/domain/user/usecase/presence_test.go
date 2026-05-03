package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"air-social/internal/domain/user/usecase"
	ucmocks "air-social/internal/domain/user/usecase/mocks"
	"air-social/pkg"
)

type presenceUseCaseSuite struct {
	suite.Suite
}

func TestPresenceUseCaseSuite(t *testing.T) {
	suite.Run(t, new(presenceUseCaseSuite))
}

type presenceUCDeps struct {
	store *ucmocks.MockPresenceStore
}

func newPresenceUCDeps(t interface {
	mock.TestingT
	Cleanup(func())
}) presenceUCDeps {
	return presenceUCDeps{
		store: ucmocks.NewMockPresenceStore(t),
	}
}

func (d presenceUCDeps) newUC() *usecase.PresenceUseCase {
	return usecase.NewPresenceUseCase(usecase.PresenceDeps{Store: d.store})
}

func (s *presenceUseCaseSuite) TestGetBatchStatus() {
	ctx := context.Background()

	tests := []struct {
		name      string
		userIDs   []int64
		setupMock func(d presenceUCDeps)
		wantErr   error
		wantResult map[int64]bool
	}{
		{
			name:    "success_mixed",
			userIDs: []int64{1, 2, 3},
			setupMock: func(d presenceUCDeps) {
				d.store.EXPECT().IsOnlineBatch(ctx, []int64{1, 2, 3}).
					Return(map[int64]bool{1: true, 2: false, 3: true}, nil).Once()
			},
			wantResult: map[int64]bool{1: true, 2: false, 3: true},
		},
		{
			name:    "success_empty_input",
			userIDs: []int64{},
			setupMock: func(d presenceUCDeps) {
				d.store.EXPECT().IsOnlineBatch(ctx, []int64{}).
					Return(map[int64]bool{}, nil).Once()
			},
			wantResult: map[int64]bool{},
		},
		{
			name:    "store_error",
			userIDs: []int64{1},
			setupMock: func(d presenceUCDeps) {
				d.store.EXPECT().IsOnlineBatch(ctx, []int64{1}).
					Return(nil, pkg.ErrInternal).Once()
			},
			wantErr: pkg.ErrInternal,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			d := newPresenceUCDeps(s.T())
			tt.setupMock(d)

			result, err := d.newUC().GetBatchStatus(ctx, tt.userIDs)

			if tt.wantErr != nil {
				s.ErrorIs(err, tt.wantErr)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.Equal(tt.wantResult, result)
			}
		})
	}
}
