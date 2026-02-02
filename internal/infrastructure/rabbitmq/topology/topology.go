package topology

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/infrastructure/rabbitmq/config"
)

func SetupExchange(ch *amqp.Channel, cfg config.ExchangeConfig) error {
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

func SetupQueue(ch *amqp.Channel, cfg config.QueueConfig) (string, error) {
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

func declareAndBindDLQ(ch *amqp.Channel, cfg config.QueueConfig) error {
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

func BindQueue(
	ch *amqp.Channel,
	queue string,
	eCfg config.ExchangeConfig,
	qCfg config.QueueConfig,
) error {
	return ch.QueueBind(
		queue,
		qCfg.RoutingKey,
		eCfg.Name,
		false,
		nil,
	)
}

func SetupQos(ch *amqp.Channel) error {
	return ch.Qos(1, 0, false)
}

func StartConsume(
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
