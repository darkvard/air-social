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
