package post_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	cachemocks "air-social/internal/cache/mocks"
	"air-social/internal/domain/post"
	"air-social/pkg"
)

type postCacheSuite struct {
	suite.Suite
}

func TestPostCacheSuite(t *testing.T) {
	suite.Run(t, new(postCacheSuite))
}

// --- Get ---

func (s *postCacheSuite) TestGet_StoreHit() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	cached := post.Post{ID: 1, Content: "cached"}
	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:1", mock.Anything).
		Return(cached, nil).Once()

	got, err := c.Get(context.Background(), 1, func(context.Context) (*post.Post, error) {
		s.Fail("loader must not be called on store hit")
		return nil, nil
	})

	s.NoError(err)
	s.Equal(int64(1), got.ID)
	s.Equal("cached", got.Content)
}

func (s *postCacheSuite) TestGet_StoreMiss_CallsLoader() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:2", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (post.Post, error)) (post.Post, error) {
			return loader(ctx)
		}).Once()

	got, err := c.Get(context.Background(), 2, func(context.Context) (*post.Post, error) {
		return &post.Post{ID: 2, Content: "from db"}, nil
	})

	s.NoError(err)
	s.Equal(int64(2), got.ID)
	s.Equal("from db", got.Content)
}

func (s *postCacheSuite) TestGet_LoaderError_Propagates() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:3", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (post.Post, error)) (post.Post, error) {
			return loader(ctx)
		}).Once()

	got, err := c.Get(context.Background(), 3, func(context.Context) (*post.Post, error) {
		return nil, pkg.ErrNotFound
	})

	s.Nil(got)
	s.ErrorIs(err, pkg.ErrNotFound)
}

func (s *postCacheSuite) TestGet_ReturnsCopy_MutationSafe() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	cached := post.Post{ID: 5, Content: "original"}
	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:5", mock.Anything).
		Return(cached, nil).Times(2)

	got1, _ := c.Get(context.Background(), 5, nil)
	got1.Content = "mutated by caller"

	got2, _ := c.Get(context.Background(), 5, nil)
	s.Equal("original", got2.Content, "mutation of first result must not affect subsequent cache hits")
}

// --- GetBatch ---

func (s *postCacheSuite) TestGetBatch_AllHit() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().GetOrLoad(mock.Anything, "post:core:10", mock.Anything).Return(post.Post{ID: 10}, nil).Once()
	store.EXPECT().GetOrLoad(mock.Anything, "post:core:20", mock.Anything).Return(post.Post{ID: 20}, nil).Once()

	got, err := c.GetBatch(context.Background(), []int64{10, 20},
		func(context.Context, []int64) ([]*post.Post, error) {
			s.Fail("batchLoader must not be called when all IDs hit cache")
			return nil, nil
		},
	)

	s.NoError(err)
	s.Len(got, 2)
	s.ElementsMatch([]int64{10, 20}, []int64{got[0].ID, got[1].ID})
}

func (s *postCacheSuite) TestGetBatch_AllMiss_CallsBatchLoader() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	callLoader := func(ctx context.Context, _ string, loader func(context.Context) (post.Post, error)) (post.Post, error) {
		return loader(ctx)
	}
	store.EXPECT().GetOrLoad(mock.Anything, "post:core:10", mock.Anything).RunAndReturn(callLoader).Once()
	store.EXPECT().GetOrLoad(mock.Anything, "post:core:20", mock.Anything).RunAndReturn(callLoader).Once()
	store.EXPECT().Set(mock.Anything, "post:core:10", mock.Anything).Return(nil).Once()
	store.EXPECT().Set(mock.Anything, "post:core:20", mock.Anything).Return(nil).Once()

	got, err := c.GetBatch(context.Background(), []int64{10, 20},
		func(_ context.Context, ids []int64) ([]*post.Post, error) {
			s.ElementsMatch([]int64{10, 20}, ids)
			return []*post.Post{{ID: 10}, {ID: 20}}, nil
		},
	)

	s.NoError(err)
	s.Len(got, 2)
}

func (s *postCacheSuite) TestGetBatch_PartialMiss() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().GetOrLoad(mock.Anything, "post:core:10", mock.Anything).Return(post.Post{ID: 10}, nil).Once()
	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:20", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (post.Post, error)) (post.Post, error) {
			return loader(ctx)
		}).Once()
	store.EXPECT().Set(mock.Anything, "post:core:20", mock.Anything).Return(nil).Once()

	got, err := c.GetBatch(context.Background(), []int64{10, 20},
		func(_ context.Context, ids []int64) ([]*post.Post, error) {
			assert.Equal(s.T(), []int64{20}, ids, "DB must be called only for missed IDs")
			return []*post.Post{{ID: 20}}, nil
		},
	)

	s.NoError(err)
	s.Len(got, 2)
}

func (s *postCacheSuite) TestGetBatch_DeletedPostSkipped() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().
		GetOrLoad(mock.Anything, "post:core:99", mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, loader func(context.Context) (post.Post, error)) (post.Post, error) {
			return loader(ctx)
		}).Once()

	got, err := c.GetBatch(context.Background(), []int64{99},
		func(context.Context, []int64) ([]*post.Post, error) {
			return []*post.Post{}, nil
		},
	)

	s.NoError(err)
	s.Len(got, 0, "deleted post must be silently skipped")
}

// --- Invalidate ---

func (s *postCacheSuite) TestInvalidate() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().Invalidate(mock.Anything, "post:core:7").Return(nil).Once()

	s.NoError(c.Invalidate(context.Background(), 7))
}

func (s *postCacheSuite) TestInvalidate_StoreError_Propagates() {
	store := cachemocks.NewMockTieredStore[post.Post](s.T())
	c := post.NewCache(store)

	store.EXPECT().Invalidate(mock.Anything, "post:core:7").Return(assert.AnError).Once()

	s.ErrorIs(c.Invalidate(context.Background(), 7), assert.AnError)
}
