package config

import (
	"fmt"
	"time"
)

type PostgresConfig struct {
	DSN         string
	MaxIdleConn int
	MaxOpenConn int
	MaxLifeTime time.Duration
	MaxIdleTime time.Duration
}

func PostgresCfg() PostgresConfig {
	return PostgresConfig{
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
	}
}
