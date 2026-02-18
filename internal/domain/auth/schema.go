package auth

import "time"

type LoginParams struct {
	Email    string
	Password string
	DeviceID string
}

type RegisterParams struct {
	Email    string
	Username string
	Password string
}

type ResetPasswordParams struct {
	EmailToken string
	Password   string
}

type LogoutParams struct {
	UserID       int64
	DeviceID     string
	IsAllDevices bool
	Token        string
	ExpiresAt    time.Time
}

type TokenResult struct {
	Type            string
	AccessToken     string
	RefreshToken    string
	AccessExpireAt  time.Time
	RefreshExpireAt time.Time
}
