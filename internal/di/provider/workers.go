package provider

import (
	"air-social/internal/domain/shared"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
)

func NewWorkers(infra *Infrastructure, adapters Adapter) *worker.Manager {
	deps := email.Deps{
		Conn:       infra.Rabbit,
		Cache:      adapters.Cache,
		Dispatcher: email.NewDispatcher(adapters.Mailer),
	}
	return worker.NewManager(
		email.NewConsumer(deps, shared.EventVerify),
		email.NewConsumer(deps, shared.EventResetPassword),
	)
}
