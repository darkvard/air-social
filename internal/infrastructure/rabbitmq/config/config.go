package config

import "air-social/internal/domain"

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
	RoutingKey:           domain.RoutingKeyEmailVerify,
	DeadLetterExchange:   EventsExchange.Name,
	DeadLetterQueue:      "email_verify_queue.dlq",
	DeadLetterRoutingKey: domain.RoutingKeyEmailVerify + ".dlq",
}

var EmailResetPasswordQueueConfig = QueueConfig{
	Queue:                "email_reset_password_queue",
	RoutingKey:           domain.RoutingKeyEmailResetPassword,
	DeadLetterExchange:   EventsExchange.Name,
	DeadLetterQueue:      "email_reset_password_queue.dlq",
	DeadLetterRoutingKey: domain.RoutingKeyEmailResetPassword + ".dlq",
}
