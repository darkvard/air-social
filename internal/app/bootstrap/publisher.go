package bootstrap

import (
	amqp "github.com/rabbitmq/amqp091-go"

	mess "air-social/internal/infrastructure/messaging"
)

func NewPublisher(conn *amqp.Connection) *mess.Publisher {
	pub, err := mess.NewPublisher(
		conn,
		mess.EventsExchange,
		10,
	)
	if err != nil {
		panic(err)
	}
	return pub
}
