package topology

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"air-social/internal/infrastructure/rabbitmq/config"
)

// SetupExchange defines the message routing hub
func SetupExchange(ch *amqp.Channel, cfg config.ExchangeConfig) error {
	return ch.ExchangeDeclare(
		cfg.Name,
		cfg.Type,
		true,  // durable: survive broker restart
		false, // auto-deleted: no, keep it even if not used
		false, // internal: no, can be published to directly
		false, // no-wait: no, wait for broker confirmation
		nil,
	)
}

// SetupQueue defines the storage for messages
func SetupQueue(ch *amqp.Channel, cfg config.QueueConfig) (string, error) {
	args := amqp.Table{}
	// setup Dead Letter Exchange (DLX) for failed/rejected messages
	if cfg.DeadLetterExchange != "" && cfg.DeadLetterRoutingKey != "" {
		args["x-dead-letter-exchange"] = cfg.DeadLetterExchange
		args["x-dead-letter-routing-key"] = cfg.DeadLetterRoutingKey
	}

	q, err := ch.QueueDeclare(
		cfg.Queue,
		true,  // durable: queue survives broker restart
		false, // auto-delete: no
		false, // exclusive: no, multiple consumers allowed
		false, // no-wait: no
		args,
	)
	if err != nil {
		return "", err
	}

	// declare the DLQ (garbage bin for failed tasks) if configured
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

// SetupQos limits the number of unacknowledged messages on this channel
func SetupQos(ch *amqp.Channel) error {
	// prefetch count = 1: the worker only takes ONE message at a time.
	// It won't get a new one until the current one is Acked or Nacked.
	// This provides backpressure and fair dispatching.
	return ch.Qos(1, 0, false)
}
