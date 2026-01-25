package worker

import (
	"context"
	"sync"
)

type Worker interface {
	Start(ctx context.Context, wg *sync.WaitGroup) error
	Stop() error
}

type Manager struct {
	workers []Worker
	wg      sync.WaitGroup
}

func NewManager(workers ...Worker) *Manager {
	return &Manager{workers: workers}
}

func (m *Manager) Start(ctx context.Context) error {
	for _, w := range m.workers {
		if err := w.Start(ctx, &m.wg); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	for _, w := range m.workers {
		_ = w.Stop()
	}

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
