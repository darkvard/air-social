package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Token    TokenConfig
	Mailer   MailConfig
	RabbitMQ RabbitMQConfig
	MinIO    FileStorageConfig
	Limiter  RateLimiterConfig
}

func Load() *Config {
	envFile := ".env"
	if os.Getenv("APP_ENV") == "debug" {
		envFile = ".env.local"
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Fatal("No .env file found")
	}

	return &Config{
		Server:   ServerCfg(),
		Postgres: PostgresCfg(),
		Redis:    RedisCfg(),
		Token:    TokenCfg(),
		Mailer:   MailCfg(),
		RabbitMQ: RabbitMQCfg(),
		MinIO:    LoadFileStorageConfig(),
		Limiter:  RateLimiterConfig{},
	}
}

func getString(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func getBool(k string, d bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}
	return b
}

func getInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

func getDuration(k string, d time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	t, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return t
}
