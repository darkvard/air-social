package http

const (
	Health     = "/health"
	SwaggerAny = "/swagger/*any"
)

const (
	AuthGroup      = "/auth"
	Register       = "/register"
	Login          = "/login"
	Refresh        = "/refresh"
	ResetPassword  = "/reset-password"
	ForgotPassword = "/forgot-password"
	VerifyEmail    = "/verify-email"
	Logout         = "/logout"
)

const (
	UserGroup = "/users"
	Me        = "/me"
	Password  = "/password"
)

const (
	FileGroup = "file"
	Presigned = "/presigned"
	Confirm   = "/confirm"
)
