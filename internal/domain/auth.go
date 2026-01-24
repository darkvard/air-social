package domain

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

type RefreshRequest struct {
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
	User  UserResponse `json:"user"`
	Token TokenInfo    `json:"token"`
}

type AuthClaims struct {
	UserID   int64
	DeviceID string
	Role     int64
}

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

type LogoutParams struct {
	UserID       int64
	DeviceID     string
	IsAllDevices bool
}

type ResetPasswordParams struct {
	EmailToken string
	Password   string
}
