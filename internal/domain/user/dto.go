package user

type AuthenticateParams struct {
	Email    string
	Password string
}

type CreateParams struct {
	Email       string
	Username    string
	NewPassword string
}

type UpdateParams struct {
	UserID   int64
	FullName *string
	Bio      *string
	Location *string
	Website  *string
	Username *string
}

type ChangePasswordParams struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}

type ResetPasswordParams struct {
	Email       string
	NewPassword string
}

type UserSummaryResult struct {
	ID         int64
	FullName   string
	Bio        string
	Avatar     string
	CoverImage string
	Verified   bool
}
