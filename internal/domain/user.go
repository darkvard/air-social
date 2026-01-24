package domain

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
	// Identifier
	ID           int64  `db:"id" json:"id"`
	Email        string `db:"email" json:"email"`
	Username     string `db:"username" json:"username"`
	PasswordHash string `db:"password_hash" json:"-"`

	// Profile
	Profile

	// System info
	Verified   bool       `db:"verified" json:"verified"`
	VerifiedAt *time.Time `db:"verified_at" json:"verified_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
	Version    int        `db:"version" json:"version"`
}

type Profile struct {
	FullName   string `db:"full_name" json:"full_name"`
	Bio        string `db:"bio" json:"bio"`
	Avatar     string `db:"avatar" json:"avatar"`
	CoverImage string `db:"cover_image" json:"cover_image"`
	Location   string `db:"location" json:"location"`
	Website    string `db:"website" json:"website"`
}

type UpdateProfileRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,min=2,max=100"`
	Bio      *string `json:"bio" binding:"omitempty,max=255"`
	Location *string `json:"location" binding:"omitempty,max=100"`
	Website  *string `json:"website" binding:"omitempty,max=255"`
	Username *string `json:"username" binding:"omitempty,alphanum,min=3,max=50"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=64"`
}

type ConfirmProfileImageRequest struct {
	ObjectKey string        `json:"object_key" binding:"required"`
	Domain    UploadDomain  `json:"domain" binding:"required,oneof=users"`
	Feature   UploadFeature `json:"feature" binding:"required,oneof=avatar cover"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	Profile
}

type CreateUserParams struct {
	Email        string
	Username     string
	PasswordHash string
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

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Profile:   u.Profile,
		Verified:  u.Verified,
		CreatedAt: u.CreatedAt,
	}
}
