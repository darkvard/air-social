package config

type ServerConfig struct {
	Env          string
	AppName      string
	Protocol     string
	Domain       string
	Version      string
	Port         string
	AuthUsername string
	AuthPassword string
}

func ServerCfg() ServerConfig {
	return ServerConfig{
		Env:          getString("APP_ENV", "development"),
		AppName:      getString("APP_NAME", "air-social"),
		Protocol:     getString("APP_PROTOCOL", "http"),
		Domain:       getString("APP_DOMAIN", "localhost"),
		Version:      getString("APP_VERSION", "v1"),
		Port:         getString("APP_PORT", "8080"),
		AuthUsername: getString("APP_BASIC_AUTH_USERNAME", "admin"),
		AuthPassword: getString("APP_BASIC_AUTH_PASSWORD", "password"),
	}
}
