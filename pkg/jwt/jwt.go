package jwt

import "time"

type JWTAuthenticator struct {
	Secret          string
	Aud             string
	Iss             string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewJWTAuthenticator(secret, aud, iss string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		Secret:          secret,
		Aud:             aud,
		Iss:             iss,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}
