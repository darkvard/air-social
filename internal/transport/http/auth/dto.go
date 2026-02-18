package auth

import (
	"time"

	"air-social/internal/domain/auth"
	"air-social/internal/transport/http/user"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	DeviceID string `json:"device_id" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	IsAllDevices bool `json:"is_all_devices,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type LoginResponse struct {
	User  user.UserDetailResponse `json:"user"`
	Token TokenResponse           `json:"token"`
}

type TokenResponse struct {
	Type            string    `json:"type"`
	AccessToken     string    `json:"access_token"`
	AccessExpireAt  time.Time `json:"access_expires_at"`
	RefreshToken    string    `json:"refresh_token"`
	RefreshExpireAt time.Time `json:"refresh_expires_at"`
}

func NewTokenResponse(t auth.TokenResult) TokenResponse {
	return TokenResponse{
		Type:            t.Type,
		AccessToken:     t.AccessToken,
		RefreshToken:    t.RefreshToken,
		AccessExpireAt:  t.AccessExpireAt,
		RefreshExpireAt: t.RefreshExpireAt,
	}
}
