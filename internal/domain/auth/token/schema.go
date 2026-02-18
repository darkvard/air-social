package token

import "time"

type AccessTokenResult struct {
	Token     string
	ExpiresAt time.Time
}

type RefreshTokenResult struct {
	Raw       string
	Hashed    string
	ExpiresAt time.Time
}