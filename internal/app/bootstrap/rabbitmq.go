package bootstrap

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/config"
)

func NewRabbitMQ(cfg config.RabbitMQConfig) *amqp.Connection {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		panic(err)
	}
	return conn
}
