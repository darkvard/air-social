package config

import "time"

type ServerConfig struct {
	Env          string
	AppName      string
	Protocol     string
	Domain       string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Version      string
	AuthUsername string
	AuthPassword string
}

func ServerCfg() ServerConfig {
	return ServerConfig{
		Env:          getString("SERVER_ENV", "development"),
		AppName:      getString("SERVER_NAME", "air-social"),
		Protocol:     getString("SERVER_PROTOCOL", "http"),
		Domain:       getString("SERVER_DOMAIN", "localhost"),
		Port:         getString("SERVER_PORT", "8080"),
		Version:      getString("SERVER_VERSION", "v1"),
		ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		AuthUsername: getString("SERVER_BASIC_AUTH_USERNAME", "admin"),
		AuthPassword: getString("SERVER_BASIC_AUTH_PASSWORD", "password"),
	}
}
