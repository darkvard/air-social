package domain
// todo: remove

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	UpdateProfileImages(ctx context.Context, userID int64, url string, feature UploadFeature) error
}

type User struct {
	ID           int64
	Email        string
	Username     string
	PasswordHash string
	Profile      Profile
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      int
}

type UserStatus struct {
	Verified   bool
	VerifiedAt *time.Time
	// Role int
}

type Profile struct {
	FullName   string
	Bio        string
	Avatar     string
	CoverImage string
	Location   string
	Website    string
}

type UserSummary struct {
	ID         int64
	FullName   string
	Bio        string
	Avatar     string
	CoverImage string
	Verified   bool
}

type CreateUserParams struct {
	Email          string
	Username       string
	PasswordHashed string
}

type UpdateProfileParams struct {
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
