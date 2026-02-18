package topology

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/infrastructure/rabbitmq/config"
)

// PrepareConsumerChannel orchestrates the full setup for a reliable consumer
func PrepareConsumerChannel(conn *amqp.Connection, eCfg config.ExchangeConfig, qCfg config.QueueConfig) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// cleanup helper: close channel if any setup step fails
	fail := func(err error) (*amqp.Channel, error) {
		ch.Close()
		return nil, err
	}

	// 1. ensure Exchange exists
	if err := SetupExchange(ch, eCfg); err != nil {
		return fail(err)
	}

	// 2. ensure Queue exists (and its DLX)
	queueName, err := SetupQueue(ch, qCfg)
	if err != nil {
		return fail(err)
	}

	// 3. bind Queue to Exchange with a specific Routing Key
	if err := BindQueue(ch, queueName, eCfg, qCfg); err != nil {
		return fail(err)
	}

	// 4. set Prefetch to 1 (Process one-by-one mode)
	if err := SetupQos(ch); err != nil {
		return fail(err)
	}
	return ch, nil
}

// StartConsume opens the door for messages to flow in
func StartConsume(ch *amqp.Channel, queue string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queue,
		"",    // consumer tag: auto-generated
		false, // auto-ack: no, we will manually Ack/Nack after processing
		false, // exclusive: no
		false, // no-local: no
		false, // no-wait: no
		nil,
	)
}
