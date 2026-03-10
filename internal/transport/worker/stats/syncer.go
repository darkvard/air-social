package stats

import (
	"context"
	"sync"
	"time"

	domainStats "air-social/internal/domain/stats"
	"air-social/pkg"
)

type Syncer struct {
	usecase domainStats.UseCase
	ticker  *time.Ticker
	done    chan struct{}
	once    sync.Once
}

func NewSyncer(uc domainStats.UseCase, interval time.Duration) *Syncer {
	return &Syncer{
		usecase: uc,
		ticker:  time.NewTicker(interval),
		done:    make(chan struct{}),
	}
}

func (s *Syncer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			case <-s.ticker.C:
				syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

				if err := s.usecase.SyncPostStats(syncCtx); err != nil {
					pkg.Log().Error("failed to sync post stats", err)
				}

				if err := s.usecase.SyncCommentStats(syncCtx); err != nil {
					pkg.Log().Error("failed to sync comment stats", err)
				}

				cancel()
			}
		}
	})
	return nil
}

func (s *Syncer) Stop() error {
	s.once.Do(func() {
		s.ticker.Stop()
		close(s.done)
	})
	return nil
}
