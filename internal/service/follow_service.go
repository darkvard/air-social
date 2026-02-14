package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"air-social/internal/domain"
	"air-social/pkg"
)

type FollowService interface {
	Follow(ctx context.Context, followerID, followeeID int64) error
	Unfollow(ctx context.Context, followerID, followeeID int64) error
	GetFollowings(ctx context.Context, params domain.FollowParams) (domain.PaginatedResult[domain.SocialUser], error)
	GetFollowers(ctx context.Context, params domain.FollowParams) (domain.PaginatedResult[domain.SocialUser], error)
}

type FollowServiceImpl struct {
	followRepo domain.FollowRepository
	cache      domain.CacheStorage
	userSvc    UserService
}

func NewFollowService(followRepo domain.FollowRepository, cache domain.CacheStorage, userSvc UserService) *FollowServiceImpl {
	return &FollowServiceImpl{
		followRepo: followRepo,
		cache:      cache,
		userSvc:    userSvc,
	}
}

func (s *FollowServiceImpl) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("cannot follow yourself: %w", pkg.ErrBadRequest)
	}

	user, err := s.userSvc.GetSummary(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found: %w", pkg.ErrBadRequest)
	}

	if err := s.followRepo.Create(ctx, &domain.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}); err != nil {
		return pkg.OrInternalError(err)
	}

	go s.invalidateFollowCache(context.Background(), followerID, followeeID)
	return nil
}

func (s *FollowServiceImpl) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("cannot unfollow yourself: %w", pkg.ErrBadRequest)
	}

	if err := s.followRepo.Delete(ctx, followerID, followeeID); err != nil {
		return pkg.OrInternalError(err)
	}

	go s.invalidateFollowCache(context.Background(), followerID, followeeID)
	return nil
}

func (s *FollowServiceImpl) GetFollowings(ctx context.Context, params domain.FollowParams) (domain.PaginatedResult[domain.SocialUser], error) {
	params.EnsureDefaults()

	users, total, err := s.runConcurrentFetch(ctx,
		func(ctx context.Context) ([]domain.SocialUser, error) {
			res, err := s.followRepo.GetFollowings(ctx, params)
			if err != nil {
				return nil, pkg.OrInternalError(err)
			}
			return s.enrichSocialUsers(ctx, params.CurrentUserID, res)
		},
		func(ctx context.Context) (int64, error) {
			return s.fetchTotalCount(ctx, params.TargetUserID, domain.GetFollowingCountKey, s.followRepo.CountFollowings)
		})

	if err != nil {
		return domain.PaginatedResult[domain.SocialUser]{}, fmt.Errorf("get following failed: %w", pkg.OrInternalError(err))
	}

	return domain.NewPaginatedResult(users, total, params.Page, params.Limit), nil
}

func (s *FollowServiceImpl) GetFollowers(ctx context.Context, params domain.FollowParams) (domain.PaginatedResult[domain.SocialUser], error) {
	params.EnsureDefaults()

	users, total, err := s.runConcurrentFetch(ctx,
		func(c context.Context) ([]domain.SocialUser, error) {
			res, err := s.followRepo.GetFollowers(c, params)
			if err != nil {
				return nil, pkg.OrInternalError(err)
			}
			return s.enrichSocialUsers(c, params.CurrentUserID, res)
		},
		func(c context.Context) (int64, error) {
			return s.fetchTotalCount(c, params.TargetUserID, domain.GetFollowerCountKey, s.followRepo.CountFollowers)
		},
	)

	if err != nil {
		return domain.PaginatedResult[domain.SocialUser]{}, fmt.Errorf("get followers failed: %w", pkg.OrInternalError(err))
	}

	return domain.NewPaginatedResult(users, total, params.Page, params.Limit), nil
}

// Internal helper

func (s *FollowServiceImpl) runConcurrentFetch(
	ctx context.Context,
	fetchFn func(ctx context.Context) ([]domain.SocialUser, error),
	countFn func(ctx context.Context) (int64, error),
) ([]domain.SocialUser, int64, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		users    []domain.SocialUser
		total    int64
		errFinal error
		wg       sync.WaitGroup
		errOnce  sync.Once
	)

	fail := func(err error) {
		if err == nil {
			return
		}
		errOnce.Do(func() {
			errFinal = err
			cancel()
		})
	}

	wg.Add(2)

	// Task 1: List
	go func() {
		defer wg.Done()
		res, err := fetchFn(ctx)
		if err != nil {
			fail(err)
			return
		}
		users = res
	}()

	// Task 2: Count
	go func() {
		defer wg.Done()
		res, err := countFn(ctx)
		if err != nil {
			fail(err)
			return
		}
		total = res
	}()

	wg.Wait()

	return users, total, errFinal
}

func (s *FollowServiceImpl) enrichSocialUsers(ctx context.Context, currentUserID int64, users []domain.User) ([]domain.SocialUser, error) {
	size := len(users)
	result := make([]domain.SocialUser, size)
	targetIDs := make([]int64, size)

	for i, u := range users {
		result[i] = domain.SocialUser{User: u}
		targetIDs[i] = u.ID
	}

	if currentUserID <= 0 || size == 0 {
		return result, nil
	}

	// Run sequentially to prevent nested concurrency and DB connection pool exhaustion.
	// Parent already runs concurrently; adding goroutines here wastes CPU on fast queries.
	followingMap, err := s.followRepo.IsFollowing(ctx, currentUserID, targetIDs)
	if err != nil {
		return nil, err
	}

	followedByMap, err := s.followRepo.IsFollowedBy(ctx, currentUserID, targetIDs)
	if err != nil {
		return nil, err
	}

	for i, u := range users {
		if followingMap[u.ID] {
			result[i].Relation.IsFollowing = true
		}
		if followedByMap[u.ID] {
			result[i].Relation.IsFollowedBy = true
		}
	}

	return result, nil
}

func (s *FollowServiceImpl) fetchTotalCount(
	ctx context.Context,
	userID int64,
	keyFunc func(int64) string,
	dbCount func(context.Context, int64) (int64, error),
) (int64, error) {
	key := keyFunc(userID)
	var total int64

	if err := s.cache.Get(ctx, key, &total); err == nil {
		return total, nil
	}

	total, err := dbCount(ctx, userID)
	if err != nil {
		return 0, err
	}

	go func(t int64) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.cache.Set(bgCtx, key, t, time.Hour)
	}(total)

	return total, nil
}

func (s *FollowServiceImpl) invalidateFollowCache(ctx context.Context, followerID, followeeID int64) {
	s.cache.Delete(ctx, domain.GetFollowingCountKey(followerID))
	s.cache.Delete(ctx, domain.GetFollowerCountKey(followeeID))
}
