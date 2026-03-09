package provider

import (
	"air-social/internal/infrastructure/rabbitmq/consumer"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
	"air-social/internal/transport/worker/stats"
)

func NewWorkers(infra *Infrastructure, adapter Adapter) *worker.Manager {
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
		Dispatcher: stats.NewDispatcher(nil),		// todo inject
	})

	return worker.NewManager(emailWorker, statsWorker)
}
