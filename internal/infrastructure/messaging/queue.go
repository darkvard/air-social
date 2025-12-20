package messaging

type QueueConfig struct {
	Queue      string
	RoutingKey string
}

var EmailRegisterQueueConfig = QueueConfig{
	Queue:      "email_register_queue",
	RoutingKey: "email.register",
}
