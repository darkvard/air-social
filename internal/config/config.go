package config

import (
	"fmt"
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
	Limiter  RateLimiterConfig
}

type ServerConfig struct {
	Port     string
	HostPort string
}

type PostgresConfig struct {
	DSN         string
	MaxIdleConn int
	MaxOpenConn int
	MaxLifeTime time.Duration
	MaxIdleTime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type TokenConfig struct {
	Secret          string
	Aud             string
	Iss             string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// todo: mailer, limiter
type MailConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

type RateLimiterConfig struct {
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
		AppEnv: getString("APP_ENV", "development"),
		Server: ServerConfig{
			Port:     getString("APP_PORT", "8080"),
			HostPort: getString("APP_HOST_PORT", "3000"),
		},
		Postgres: PostgresConfig{
			DSN: func() string {
				user := getString("DB_USER", "postgres")
				pass := getString("DB_PASS", "postgres")
				name := getString("DB_NAME", "air_social")
				host := getString("DB_HOST", "db")
				port := getString("DB_PORT", "5432")
				sslMode := getString("DB_SSL_MODE", "disable")
				dsn := fmt.Sprintf(
					"postgres://%s:%s@%s:%s/%s?sslmode=%s",
					user, pass, host, port, name, sslMode,
				)
				return getString("DB_DSN", dsn)
			}(),
			MaxIdleConn: getInt("DB_MAX_IDLE", 5),
			MaxOpenConn: getInt("DB_MAX_OPEN", 10),
			MaxLifeTime: getDuration("DB_MAX_LIFETIME", time.Hour),
			MaxIdleTime: getDuration("DB_MAX_IDLE_TIME", time.Minute*15),
		},
		Redis: RedisConfig{
			Addr: func() string {
				host := getString("REDIS_HOST", "redis")
				port := getString("REDIS_PORT", "6379")
				return fmt.Sprintf("%s:%s", host, port)
			}(),
			Password: getString("REDIS_PASS", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		Token: TokenConfig{
			Secret:          getString("JWT_SECRET", "my_secret_key"),
			Aud:             getString("JWT_AUD", "air-social"),
			Iss:             getString("JWT_ISS", "air-social-api"),
			AccessTokenTTL:  getDuration("JWT_ACCESS_TTL", time.Minute*15),
			RefreshTokenTTL: getDuration("JWT_REFRESH_TTL", time.Hour*24*7),
		},
		Mailer: MailConfig{
			Host:        getString("MAILTRAP_HOST", "smtp.mailtrap.io"),
			Port:        getInt("MAILTRAP_PORT", 587),
			Username:    getString("MAILTRAP_USERNAME", "fd2dfd6851a016"),
			Password:    getString("MAILTRAP_PASSWORD", ""),
			FromAddress: getString("MAILTRAP_FROM_ADDRESS", "no-reply@airsocial.com"),
			FromName:    getString("MAILTRAP_FROM_NAME", "Air Social"),
		},
		Limiter: RateLimiterConfig{},
	}
}

func getString(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
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
