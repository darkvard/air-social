package worker

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/event"
	"air-social/internal/infra/mailer"
	"air-social/internal/infra/msg"
	"air-social/internal/worker/email"
)

type Worker interface {
	Start(ctx context.Context, wg *sync.WaitGroup) error
	Stop() error
}

func NewWorker(conn *amqp.Connection, cache domain.CacheStorage, cfg config.MailConfig) *Manager {
	sender := mailer.NewMailtrap(cfg)
	handler := event.NewEmailHandler(sender)
	verifyWorker := email.NewEmailWorker(conn, cache, msg.EventsExchange, msg.EmailVerifyQueueConfig, handler)
	resetPasswordWorker := email.NewEmailWorker(conn, cache, msg.EventsExchange, msg.EmailResetPasswordQueueConfig, handler)
	return NewManager(verifyWorker, resetPasswordWorker)
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
