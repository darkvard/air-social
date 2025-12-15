package config

import "fmt"

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func RedisCfg() RedisConfig {
	return RedisConfig{
		Addr: func() string {
			host := getString("REDIS_HOST", "redis")
			port := getString("REDIS_PORT", "6379")
			return fmt.Sprintf("%s:%s", host, port)
		}(),
		Password: getString("REDIS_PASS", ""),
		DB:       getInt("REDIS_DB", 0),
	}
}
