package config

import "time"

type TokenConfig struct {
	Secret          string
	Aud             string
	Iss             string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func TokenCfg() TokenConfig {
	return TokenConfig{
		Secret:          getString("JWT_SECRET", "my_secret_key"),
		Aud:             getString("JWT_AUD", "air-social"),
		Iss:             getString("JWT_ISS", "air-social-api"),
		AccessTokenTTL:  getDuration("JWT_ACCESS_TTL", time.Minute*15),
		RefreshTokenTTL: getDuration("JWT_REFRESH_TTL", time.Hour*24*7),
	}
}
