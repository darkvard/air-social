package bootstrap

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
	"air-social/internal/event"
	"air-social/internal/infrastructure/mailer"
	mess "air-social/internal/infrastructure/messaging"
	"air-social/internal/worker"
	"air-social/internal/worker/email"
)

func NewWorkerManager(conn *amqp.Connection, mailCfg config.MailConfig) *worker.Manager {
	mailSender := mailer.NewMailtrap(mailCfg)
	emailHandler := event.NewEmailHandler(mailSender)

	emailWorker := email.NewEmailWorker(
		conn,
		mess.EventsExchange,
		mess.EmailRegisterQueueConfig,
		emailHandler,
	)

	return worker.NewManager(emailWorker)
}
