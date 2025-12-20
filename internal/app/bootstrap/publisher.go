package bootstrap

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/infrastructure/messaging"
)

func NewPublisher(conn *amqp.Connection) *messaging.Publisher {
	pub, err := messaging.NewPublisher(
		conn,
		messaging.EventsExchange,
		10,
	)
	if err != nil {
		panic(err)
	}
	return pub
}
