package route

const (
	Health     = "/health"
	SwaggerAny = "/swagger/*any"
	ID         = "/:id"
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
	UserGroup  = "/users"
	Me         = "/me"
	Password   = "/password"
	Followers  = ID + "/followers"
	Followings = ID + "/followings"
	FollowUser = ID + "/follow"
)

const (
	MediaGroup      = "/media"
	PresignedUpload = "/presigned-urls"
	Images          = "/images"
)

const (
	PostGroup = "/posts"
)
