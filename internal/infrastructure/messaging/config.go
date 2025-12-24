package messaging

type ExchangeConfig struct {
	Name string
	Type string
}

type QueueConfig struct {
	Queue      string
	RoutingKey string
}

var EventsExchange = ExchangeConfig{
	Name: "events",
	Type: "topic",
}

var EmailRegisterQueueConfig = QueueConfig{
	Queue:      "email_register_queue",
	RoutingKey: "email.register",
}