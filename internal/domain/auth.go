package domain

type AuthClaims struct {
	UserID   int64
	DeviceID string
	Role     int64
}

type TokenMeta struct {
	AccessToken string
	ExpiresAt   int64
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
	Token        TokenMeta
}

type ResetPasswordParams struct {
	EmailToken string
	Password   string
}
