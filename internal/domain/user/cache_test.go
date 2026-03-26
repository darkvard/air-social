package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachemocks "air-social/internal/cache/mocks"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type userCacheSuite struct {
	suite.Suite
}

func TestUserCacheSuite(t *testing.T) {
	suite.Run(t, new(userCacheSuite))
}

func (s *userCacheSuite) TestGet_StoreHit() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	cached := user.UserSummary{ID: 1, FullName: "Alice"}
	store.EXPECT().
		GetOrLoad(mock.Anything, "user:detail:1", mock.Anything).
		Return(cached, nil).Once()

	got, err := c.Get(context.Background(), 1, func(context.Context) (*user.UserSummary, error) {
		s.Fail("loader must not be called on store hit")
		return nil, nil
	})

	s.NoError(err)
	s.Equal(int64(1), got.ID)
	s.Equal("Alice", got.FullName)
}

func (s *userCacheSuite) TestGet_StoreMiss_CallsLoader() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	store.EXPECT().
		GetOrLoad(mock.Anything, "user:detail:2", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (user.UserSummary, error)) (user.UserSummary, error) {
			return loader(ctx)
		}).Once()

	got, err := c.Get(context.Background(), 2, func(context.Context) (*user.UserSummary, error) {
		return &user.UserSummary{ID: 2, FullName: "Bob"}, nil
	})

	s.NoError(err)
	s.Equal(int64(2), got.ID)
	s.Equal("Bob", got.FullName)
}

func (s *userCacheSuite) TestGet_LoaderError_Propagates() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	store.EXPECT().
		GetOrLoad(mock.Anything, "user:detail:3", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (user.UserSummary, error)) (user.UserSummary, error) {
			return loader(ctx)
		}).Once()

	got, err := c.Get(context.Background(), 3, func(context.Context) (*user.UserSummary, error) {
		return nil, pkg.ErrNotFound
	})

	s.Nil(got)
	s.ErrorIs(err, pkg.ErrNotFound)
}

func (s *userCacheSuite) TestGet_ReturnsCopy_MutationSafe() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	cached := user.UserSummary{ID: 5, FullName: "Original"}
	store.EXPECT().
		GetOrLoad(mock.Anything, "user:detail:5", mock.Anything).
		Return(cached, nil).Times(2)

	got1, _ := c.Get(context.Background(), 5, nil)
	got1.FullName = "Mutated"

	got2, _ := c.Get(context.Background(), 5, nil)
	s.Equal("Original", got2.FullName, "mutation must not affect subsequent cache hits")
}

func (s *userCacheSuite) TestInvalidate() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	store.EXPECT().Invalidate(mock.Anything, "user:detail:7").Return(nil).Once()

	s.NoError(c.Invalidate(context.Background(), 7))
}

func (s *userCacheSuite) TestInvalidate_StoreError_Propagates() {
	store := cachemocks.NewMockTieredStore[user.UserSummary](s.T())
	c := user.NewCache(store)

	store.EXPECT().Invalidate(mock.Anything, "user:detail:7").Return(pkg.ErrInternal).Once()

	s.ErrorIs(c.Invalidate(context.Background(), 7), pkg.ErrInternal)
}
