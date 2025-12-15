package config

import "fmt"

type RabbitMQConfig struct {
	URL string
}

func RabbitMQCfg() RabbitMQConfig {
	return RabbitMQConfig{
		URL: func() string {
			host := getString("RABBITMQ_HOST", "rabbitmq")
			port := getInt("RABBITMQ_PORT", 5672)
			_ = getInt("RABBITMQ_UI_PORT", 1567)
			user := getString("RABBITMQ_USER", "guest")
			pass := getString("RABBITMQ_PASS", "guest")
			url := fmt.Sprintf(
				"amqp://%s:%s@%s:%d",
				user, pass, host, port,
			)
			return getString("RABBITMQ_URL", url)
		}(),
	}
}
