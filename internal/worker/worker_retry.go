package worker


import amqp "github.com/rabbitmq/amqp091-go"

const (
	RetryHeaderKey = "x-retry-count"
	DefaultMaxRetry = 3
)

func GetRetryCount(msg amqp.Delivery) int {
	if msg.Headers == nil {
		return 0
	}
	if v, ok := msg.Headers[RetryHeaderKey]; ok {
		switch t := v.(type) {
		case int32:
			return int(t)
		case int:
			return t
		}
	}
	return 0
}

func IncrementRetry(msg *amqp.Delivery) {
	if msg.Headers == nil {
		msg.Headers = amqp.Table{}
	}
	msg.Headers[RetryHeaderKey] = int32(GetRetryCount(*msg) + 1)
}