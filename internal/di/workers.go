package di

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/transport/worker"
)

func initWorkers(
	infra *Infrastructures,
	adapters *Adapters,
	services *Services,
) *worker.Manager {
	exchangeCfg := config.EventsExchange

	verifyWorker := worker.NewEmailWorker(
		infra.Rabbit,
		adapters.Cache,
		services.Email,
		exchangeCfg,
		config.EmailVerifyQueueConfig,
	)

	resetWorker := worker.NewEmailWorker(
		infra.Rabbit,
		adapters.Cache,
		services.Email,
		exchangeCfg,
		config.EmailResetPasswordQueueConfig,
	)

	return worker.NewManager(verifyWorker, resetWorker)
}
