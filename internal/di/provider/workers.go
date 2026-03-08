package provider

import (
	"air-social/internal/domain/common"
	"air-social/internal/transport/worker"
	"air-social/internal/transport/worker/email"
)

func NewWorkers(infra *Infrastructure, adapter Adapter) *worker.Manager {
	deps := email.Deps{
		Conn:       infra.Rabbit,
		Cache:      adapter.Cache,
		Dispatcher: email.NewDispatcher(adapter.Mailer),
	}
	return worker.NewManager(
		email.NewConsumer(deps, common.EventEmailVerify),
		email.NewConsumer(deps, common.EventEmailResetPassword),
	)
}
