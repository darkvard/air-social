package chat

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/consumer"
)

func NewWorkerGroup(deps consumer.Deps) *consumer.Group {
	queues := []config.QueueConfig{
		config.ChatFollowCreatedQueueConfig,
	}
	return consumer.NewGroup(deps, queues, consumer.Domain("chat"))
}
