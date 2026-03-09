package stats

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/consumer"
)

func NewWorkerGroup(deps consumer.Deps) *consumer.Group {
	queues := []config.QueueConfig{
		config.SocialPostLikeQueueConfig,
		config.SocialPostShareQueueConfig,
		config.SocialCommentCreatedQueueConfig,
		config.SocialCommentDeletedQueueConfig,
		config.SocialCommentLikeQueueConfig,
	}

	return consumer.NewGroup(deps, queues, "stats")
}
