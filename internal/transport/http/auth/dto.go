package auth

import (
	"time"
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
	User  UserResponse  `json:"user"`
	Token TokenResponse `json:"token"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Verified  bool      `json:"verified"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	Type            string    `json:"type"`
	AccessToken     string    `json:"access_token"`
	AccessExpireAt  time.Time `json:"access_expires_at"`
	RefreshToken    string    `json:"refresh_token"`
	RefreshExpireAt time.Time `json:"refresh_expires_at"`
}
