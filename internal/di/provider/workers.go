package provider

import (
	"time"

	"air-social/internal/infrastructure/rabbitmq/consumer"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
	"air-social/internal/transport/worker/feed"
	"air-social/internal/transport/worker/stats"
)

func NewWorkers(infra *Infrastructure, adapter Adapter, prov Provider, usecase UseCase) *worker.Manager {
	// 1. Email Worker
	emailWorker := email.NewWorkerGroup(consumer.Deps{
		Conn:       infra.Rabbit,
		Cache:      adapter.Cache,
		Dispatcher: email.NewDispatcher(adapter.Mailer),
	})

	// 2. Stats Worker
	statsWorker := stats.NewWorkerGroup(consumer.Deps{
		Conn:       infra.Rabbit,
		Cache:      adapter.Cache,
		Dispatcher: stats.NewDispatcher(prov.Stats),
	})

	// 3. Stats Syncer
	statsSyncer := stats.NewSyncer(usecase.Stats, 1*time.Minute)

	// 4. Feed Worker
	feedWorker := feed.NewWorkerGroup(consumer.Deps{
		Conn:       infra.Rabbit,
		Cache:      adapter.Cache,
		Dispatcher: feed.NewDispatcher(usecase.Feed),
	})

	return worker.NewManager(emailWorker, statsWorker, statsSyncer, feedWorker)
}
