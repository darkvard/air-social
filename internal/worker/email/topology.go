package email

import (
	amqp "github.com/rabbitmq/amqp091-go"

	mess "air-social/internal/infrastructure/messaging"
)

func setupExchange(ch *amqp.Channel, cfg mess.ExchangeConfig) error {
	return ch.ExchangeDeclare(
		cfg.Name,
		cfg.Type,
		true, // durable
		false,
		false,
		false,
		nil,
	)
}

func setupQueue(ch *amqp.Channel, cfg mess.QueueConfig) (string, error) {
	args := amqp.Table{}
	if cfg.DeadLetterExchange != "" && cfg.DeadLetterRoutingKey != "" {
		args["x-dead-letter-exchange"] = cfg.DeadLetterExchange
		args["x-dead-letter-routing-key"] = cfg.DeadLetterRoutingKey
	}

	q, err := ch.QueueDeclare(
		cfg.Queue,
		true, // durable
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return "", err
	}

	if cfg.DeadLetterQueue != "" {
		if err := declareAndBindDLQ(ch, cfg); err != nil {
			return "", err
		}
	}

	return q.Name, nil
}

func declareAndBindDLQ(ch *amqp.Channel, cfg mess.QueueConfig) error {
	if _, err := ch.QueueDeclare(
		cfg.DeadLetterQueue,
		true, // durable
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return ch.QueueBind(
		cfg.DeadLetterQueue,
		cfg.DeadLetterRoutingKey,
		cfg.DeadLetterExchange,
		false,
		nil,
	)
}

func bindQueue(
	ch *amqp.Channel,
	queue string,
	eCfg mess.ExchangeConfig,
	qCfg mess.QueueConfig,
) error {
	return ch.QueueBind(
		queue,
		qCfg.RoutingKey,
		eCfg.Name,
		false,
		nil,
	)
}

func setupQos(ch *amqp.Channel) error {
	return ch.Qos(1, 0, false)
}

func startConsume(
	ch *amqp.Channel,
	queue string,
) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}
