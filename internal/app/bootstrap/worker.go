package bootstrap

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/event"
	"air-social/internal/infrastructure/mailer"
	mess "air-social/internal/infrastructure/messaging"
	"air-social/internal/worker"
	"air-social/internal/worker/email"
)

func NewWorkerManager(conn *amqp.Connection, c cache.CacheStorage, mCfg config.MailConfig) *worker.Manager {
	sender := mailer.NewMailtrap(mCfg)
	handler := event.NewEmailHandler(sender)
	verifyWorker := email.NewEmailWorker(conn, c, mess.EventsExchange, mess.EmailVerifyQueueConfig, handler)
	resetPasswordWorker := email.NewEmailWorker(conn, c, mess.EventsExchange, mess.EmailResetPasswordQueueConfig, handler)
	return worker.NewManager(verifyWorker, resetPasswordWorker)
}
