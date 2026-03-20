package feed

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/consumer"
)

func NewWorkerGroup(deps consumer.Deps) *consumer.Group {
	queues := []config.QueueConfig{
		config.SocialPostCreatedQueueConfig,
		config.SocialPostDeletedQueueConfig,
	}
	return consumer.NewGroup(deps, queues, consumer.Domain("feed"))
}
