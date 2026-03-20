package config

const deadExt = ".dlq"

const (
	verifyRouting          = "email.verify"
	verifyQueue            = "email_verify_queue"
	verifyQueueDead        = verifyQueue + deadExt
	verifyQueueDeadRouting = verifyRouting + deadExt
)

const (
	resetPasswordRouting          = "email.reset_password"
	resetPasswordQueue            = "email_reset_password_queue"
	resetPasswordQueueDead        = resetPasswordQueue + deadExt
	resetPasswordQueueDeadRouting = resetPasswordRouting + deadExt
)

const (
	postLikeRouting          = "social.post.like"
	postLikeQueue            = "social_post_like_queue"
	postLikeQueueDead        = postLikeQueue + deadExt
	postLikeQueueDeadRouting = postLikeRouting + deadExt
)

const (
	postShareRouting          = "social.post.share"
	postShareQueue            = "social_post_share_queue"
	postShareQueueDead        = postShareQueue + deadExt
	postShareQueueDeadRouting = postShareRouting + deadExt
)

const (
	commentLikeRouting          = "social.comment.like"
	commentLikeQueue            = "social_comment_like_queue"
	commentLikeQueueDead        = commentLikeQueue + deadExt
	commentLikeQueueDeadRouting = commentLikeRouting + deadExt
)

const (
	commentCreatedRouting          = "social.comment.created"
	commentCreatedQueue            = "social_comment_created_queue"
	commentCreatedQueueDead        = commentCreatedQueue + deadExt
	commentCreatedQueueDeadRouting = commentCreatedRouting + deadExt
)

const (
	commentDeletedRouting          = "social.comment.deleted"
	commentDeletedQueue            = "social_comment_deleted_queue"
	commentDeletedQueueDead        = commentDeletedQueue + deadExt
	commentDeletedQueueDeadRouting = commentDeletedRouting + deadExt
)

const (
	postCreatedRouting          = "social.post.created"
	postCreatedQueue            = "social_post_created_queue"
	postCreatedQueueDead        = postCreatedQueue + deadExt
	postCreatedQueueDeadRouting = postCreatedRouting + deadExt
)

const (
	postDeletedRouting          = "social.post.deleted"
	postDeletedQueue            = "social_post_deleted_queue"
	postDeletedQueueDead        = postDeletedQueue + deadExt
	postDeletedQueueDeadRouting = postDeletedRouting + deadExt
)

type ExchangeConfig struct {
	Name string
	Type string
}

type QueueConfig struct {
	Queue                string
	RoutingKey           string
	DeadLetterQueue      string
	DeadLetterRoutingKey string
	DeadLetterExchange   string
}

var TopicEventsExchange = ExchangeConfig{
	Name: "events",
	Type: "topic",
}

var EmailVerifyQueueConfig = QueueConfig{
	Queue:                verifyQueue,
	RoutingKey:           verifyRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      verifyQueueDead,
	DeadLetterRoutingKey: verifyQueueDeadRouting,
}

var EmailResetPasswordQueueConfig = QueueConfig{
	Queue:                resetPasswordQueue,
	RoutingKey:           resetPasswordRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      resetPasswordQueueDead,
	DeadLetterRoutingKey: resetPasswordQueueDeadRouting,
}

var SocialPostLikeQueueConfig = QueueConfig{
	Queue:                postLikeQueue,
	RoutingKey:           postLikeRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      postLikeQueueDead,
	DeadLetterRoutingKey: postLikeQueueDeadRouting,
}

var SocialPostShareQueueConfig = QueueConfig{
	Queue:                postShareQueue,
	RoutingKey:           postShareRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      postShareQueueDead,
	DeadLetterRoutingKey: postShareQueueDeadRouting,
}

var SocialCommentLikeQueueConfig = QueueConfig{
	Queue:                commentLikeQueue,
	RoutingKey:           commentLikeRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      commentLikeQueueDead,
	DeadLetterRoutingKey: commentLikeQueueDeadRouting,
}

var SocialCommentCreatedQueueConfig = QueueConfig{
	Queue:                commentCreatedQueue,
	RoutingKey:           commentCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      commentCreatedQueueDead,
	DeadLetterRoutingKey: commentCreatedQueueDeadRouting,
}

var SocialCommentDeletedQueueConfig = QueueConfig{
	Queue:                commentDeletedQueue,
	RoutingKey:           commentDeletedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      commentDeletedQueueDead,
	DeadLetterRoutingKey: commentDeletedQueueDeadRouting,
}

var SocialPostCreatedQueueConfig = QueueConfig{
	Queue:                postCreatedQueue,
	RoutingKey:           postCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      postCreatedQueueDead,
	DeadLetterRoutingKey: postCreatedQueueDeadRouting,
}

var SocialPostDeletedQueueConfig = QueueConfig{
	Queue:                postDeletedQueue,
	RoutingKey:           postDeletedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      postDeletedQueueDead,
	DeadLetterRoutingKey: postDeletedQueueDeadRouting,
}
