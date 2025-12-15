package config

type ServerConfig struct {
	Port     string
	HostPort string
	Env      string
}

func ServerCfg() ServerConfig {
	return ServerConfig{
		Port:     getString("APP_PORT", "8080"),
		HostPort: getString("APP_HOST_PORT", "3000"),
		Env:      getString("APP_ENV", "development"),
	}
}
