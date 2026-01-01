package messaging

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

var EventsExchange = ExchangeConfig{
	Name: "events",
	Type: "topic",
}

var EmailVerifyQueueConfig = QueueConfig{
	Queue:                "email_verify_queue",
	RoutingKey:           "email.verify",
	DeadLetterExchange:   EventsExchange.Name,
	DeadLetterQueue:      "email_verify_queue.dlq",
	DeadLetterRoutingKey: "email.verify.dlq",
}

var EmailResetPasswordQueueConfig = QueueConfig{
	Queue:                "email_reset_password_queue",
	RoutingKey:           "email.reset_password",
	DeadLetterExchange:   EventsExchange.Name,
	DeadLetterQueue:      "email_reset_password_queue.dlq",
	DeadLetterRoutingKey: "email.reset_password.dlq",
}
