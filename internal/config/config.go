package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"air-social/pkg"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Token    TokenConfig
	Mailer   MailConfig
	RabbitMQ RabbitMQConfig
	MinIO    MinioStorageConfig
	Limiter  RateLimiterCfg
}

func Load() Config {
	envFile := ".env"
	if os.Getenv("APP_ENV") == pkg.DEBUG {
		envFile = ".env.local"
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Fatal("No .env file found")
	}

	serverCfg := ServerCfg()

	return Config{
		Server:   serverCfg,
		Postgres: PostgresCfg(),
		Redis:    RedisCfg(),
		Token:    TokenCfg(),
		Mailer:   MailCfg(),
		RabbitMQ: RabbitMQCfg(),
		MinIO:    MinStorageCfg(serverCfg.AppName),
		Limiter:  RateLimiterCfg{},
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
