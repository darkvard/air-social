package config

type ServerConfig struct {
	Port     string
	HostPort string
	Env      string
	BaseURL  string
	Version  string
}

func ServerCfg() ServerConfig {
	return ServerConfig{
		Port:     getString("APP_PORT", "8080"),
		HostPort: getString("APP_HOST_PORT", "3000"),
		Env:      getString("APP_ENV", "development"),
		BaseURL:  getString("APP_BASE_URL", "http://localhost:3000"),
		Version:  getString("APP_VERSION", "v1"),
	}
}
