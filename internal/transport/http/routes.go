package http

const Health = "/health"

const (
	AuthGroup      = "/auth"
	Register       = "/register"
	Login          = "/login"
	Refresh        = "/refresh"
	ResetPassword  = "/reset-password"
	ForgotPassword = "/forgot-password"
	VerifyEmail    = "/verify-email"
	Logout         = "/logout"
	SwaggerAny     = "/swagger/*any"
)

const (
	UserGroup = "/users"
	Me        = "/me"
	Avatar    = "/avatar"
	Password  = "/password"
)