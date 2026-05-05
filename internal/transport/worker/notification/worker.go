package notification

import (
	"air-social/internal/infrastructure/rabbitmq/config"
	"air-social/internal/infrastructure/rabbitmq/consumer"
)

func NewWorkerGroup(deps consumer.Deps) *consumer.Group {
	queues := []config.QueueConfig{
		config.NotifPostLikeQueueConfig,
		config.NotifCommentLikeQueueConfig,
		config.NotifCommentCreatedQueueConfig,
		config.NotifPostShareQueueConfig,
		config.NotifFollowCreatedQueueConfig,
		config.NotifMessageCreatedQueueConfig,
	}
	return consumer.NewGroup(deps, queues, consumer.Domain("notification"))
}
