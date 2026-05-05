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
	followCreatedRouting          = "social.follow.created"
	followCreatedQueue            = "chat_follow_created_queue"
	followCreatedQueueDead        = followCreatedQueue + deadExt
	followCreatedQueueDeadRouting = followCreatedRouting + deadExt
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

var ChatFollowCreatedQueueConfig = QueueConfig{
	Queue:                followCreatedQueue,
	RoutingKey:           followCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      followCreatedQueueDead,
	DeadLetterRoutingKey: followCreatedQueueDeadRouting,
}

var SocialPostDeletedQueueConfig = QueueConfig{
	Queue:                postDeletedQueue,
	RoutingKey:           postDeletedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      postDeletedQueueDead,
	DeadLetterRoutingKey: postDeletedQueueDeadRouting,
}

// Notification worker queues.
// Topic exchange fan-out: same routing key → multiple queues receive the same message.
// Each consumer domain gets its own queue, bound to the same routing key as stats/feed.
const (
	notifPostLikeQueue           = "notif_post_like_queue"
	notifPostLikeQueueDead       = notifPostLikeQueue + deadExt
	notifCommentLikeQueue        = "notif_comment_like_queue"
	notifCommentLikeQueueDead    = notifCommentLikeQueue + deadExt
	notifCommentCreatedQueue     = "notif_comment_created_queue"
	notifCommentCreatedQueueDead = notifCommentCreatedQueue + deadExt
	notifPostShareQueue          = "notif_post_share_queue"
	notifPostShareQueueDead      = notifPostShareQueue + deadExt
	notifFollowCreatedQueue      = "notif_follow_created_queue"
	notifFollowCreatedQueueDead  = notifFollowCreatedQueue + deadExt
	notifMessageCreatedQueue     = "notif_message_created_queue"
	notifMessageCreatedQueueDead = notifMessageCreatedQueue + deadExt

	messageCreatedRouting          = "chat.message.created"
	messageCreatedQueueDeadRouting = messageCreatedRouting + deadExt
)

var NotifPostLikeQueueConfig = QueueConfig{
	Queue:                notifPostLikeQueue,
	RoutingKey:           postLikeRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifPostLikeQueueDead,
	DeadLetterRoutingKey: postLikeRouting + deadExt,
}

var NotifCommentLikeQueueConfig = QueueConfig{
	Queue:                notifCommentLikeQueue,
	RoutingKey:           commentLikeRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifCommentLikeQueueDead,
	DeadLetterRoutingKey: commentLikeRouting + deadExt,
}

var NotifCommentCreatedQueueConfig = QueueConfig{
	Queue:                notifCommentCreatedQueue,
	RoutingKey:           commentCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifCommentCreatedQueueDead,
	DeadLetterRoutingKey: commentCreatedRouting + deadExt,
}

var NotifPostShareQueueConfig = QueueConfig{
	Queue:                notifPostShareQueue,
	RoutingKey:           postShareRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifPostShareQueueDead,
	DeadLetterRoutingKey: postShareRouting + deadExt,
}

var NotifFollowCreatedQueueConfig = QueueConfig{
	Queue:                notifFollowCreatedQueue,
	RoutingKey:           followCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifFollowCreatedQueueDead,
	DeadLetterRoutingKey: followCreatedRouting + deadExt,
}

var NotifMessageCreatedQueueConfig = QueueConfig{
	Queue:                notifMessageCreatedQueue,
	RoutingKey:           messageCreatedRouting,
	DeadLetterExchange:   TopicEventsExchange.Name,
	DeadLetterQueue:      notifMessageCreatedQueueDead,
	DeadLetterRoutingKey: messageCreatedQueueDeadRouting,
}
