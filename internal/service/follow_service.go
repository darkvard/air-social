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
	GetFollowings(ctx context.Context, params domain.FollowParams) (domain.FollowResult, error)
	GetFollowers(ctx context.Context, params domain.FollowParams) (domain.FollowResult, error)
}

type FollowServiceImpl struct {
	followRepo domain.FollowRepository
	userRepo   domain.UserRepository
	cache      domain.CacheStorage
}

func NewFollowService(followRepo domain.FollowRepository, userRepo domain.UserRepository, cache domain.CacheStorage) *FollowServiceImpl {
	return &FollowServiceImpl{
		followRepo: followRepo,
		userRepo:   userRepo,
		cache:      cache,
	}
}

func (s *FollowServiceImpl) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("%w: cannot follow yourself", pkg.ErrBadRequest)
	}

	user, err := s.userRepo.GetByID(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("%w: user not found", pkg.ErrBadRequest)
	}

	if err := s.followRepo.Create(ctx, &domain.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}); err != nil {
		return pkg.OrInternalError(err)
	}

	go s.invalidateCacheAsync(context.Background(), followerID, followeeID)

	return nil
}

func (s *FollowServiceImpl) Unfollow(ctx context.Context, followerID int64, followeeID int64) error {
	if followerID == followeeID {
		return fmt.Errorf("%w: cannot unfollow yourself", pkg.ErrBadRequest)
	}

	if err := s.followRepo.Delete(ctx, followerID, followeeID); err != nil {
		return pkg.OrInternalError(err)
	}

	go s.invalidateCacheAsync(context.Background(), followerID, followeeID)

	return nil
}

func (s *FollowServiceImpl) GetFollowings(ctx context.Context, params domain.FollowParams) (domain.FollowResult, error) {
	return domain.FollowResult{}, nil
}

func (s *FollowServiceImpl) GetFollowers(ctx context.Context, params domain.FollowParams) (domain.FollowResult, error) {
	var (
		empty             domain.FollowResult
		users             []domain.User
		total             int64
		listErr, countErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		users, listErr = s.followRepo.GetFollowers(ctx, params)
	}()

	go func() {
		defer wg.Done()
		total, countErr = s.fetchTotalFollowerCount(ctx, params.UserID)
	}()

	wg.Wait()

	if listErr != nil {
		return empty, fmt.Errorf("failed to get users: %w", listErr)
	}
	if countErr != nil {
		return empty, fmt.Errorf("failed to count users: %w", countErr)
	}

	return domain.FollowResult{
		Users: users,
		Total: total,
		Page:  params.GetPage(),
		Limit: params.GetLimit(),
	}, nil
}

// Internal helper

func (s *FollowServiceImpl) fetchTotalFollowerCount(ctx context.Context, userID int64) (int64, error) {
	if cached := s.getFollowerCountCache(ctx, userID); cached > 0 {
		return cached, nil
	}

	total, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		return 0, err
	}

	go s.cache.Set(context.Background(), domain.GetFollowerCountKey(userID), total, time.Hour)

	return total, nil
}

func (s *FollowServiceImpl) invalidateCacheAsync(ctx context.Context, followerID int64, followeeID int64) {
	s.cache.Delete(ctx, domain.GetFollowingCountKey(followerID))
	s.cache.Delete(ctx, domain.GetFollowerCountKey(followeeID))
}

func (s *FollowServiceImpl) getFollowerCountCache(ctx context.Context, userID int64) int64 {
	var total int64
	key := domain.GetFollowerCountKey(userID)
	if err := s.cache.Get(ctx, key, &total); err != nil {
		return 0
	}
	return total
}

 