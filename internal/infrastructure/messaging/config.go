package messaging

type ExchangeConfig struct {
	Name string
	Type string
}

type QueueConfig struct {
	Queue      string
	RoutingKey string

	DeadLetterQueue      string
	DeadLetterRoutingKey string
	DeadLetterExchange   string
}

var EventsExchange = ExchangeConfig{
	Name: "events",
	Type: "topic",
}

var EmailVerifyQueueConfig = QueueConfig{
	Queue:      "email_verify_queue",
	RoutingKey: "email.verify",

	DeadLetterExchange:   EventsExchange.Name,
	DeadLetterQueue:      "email_verify_queue.dlq",
	DeadLetterRoutingKey: "email.verify.dlq",
}
