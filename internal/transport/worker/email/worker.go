package email

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/consumer"
)

func NewWorkerGroup(deps consumer.Deps) *consumer.Group {
	queues := []config.QueueConfig{
		config.EmailVerifyQueueConfig,
		config.EmailResetPasswordQueueConfig,
	}
	return consumer.NewGroup(deps, queues, "email")
}
