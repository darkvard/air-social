package email

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/worker"
)

func handleRetry(msg amqp.Delivery) {
	retry := worker.GetRetryCount(msg)

	if retry < worker.DefaultMaxRetry {
		worker.IncrementRetry(&msg)
		msg.Nack(false, true)
		return
	}

	msg.Nack(false, false)
}
