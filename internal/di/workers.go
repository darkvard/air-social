package di

import (
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
)

func initWorkers(
	infra *Infrastructures,
	adapters *Adapters,
	services *Services,
) *worker.Manager {
	exchangeCfg := rabbitmq.EventsExchange

	verifyWorker := email.NewEmailWorker(
		infra.Rabbit,
		adapters.Cache,
		services.Email,
		exchangeCfg,
		rabbitmq.EmailVerifyQueueConfig,
	)

	resetWorker := email.NewEmailWorker(
		infra.Rabbit,
		adapters.Cache,
		services.Email,
		exchangeCfg,
		rabbitmq.EmailResetPasswordQueueConfig,
	)

	return worker.NewManager(verifyWorker, resetWorker)
}
