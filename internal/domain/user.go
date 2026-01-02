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

type UpdateRequest struct {
	FullName *string `json:"full_name" binding:"omitempty,min=2,max=100"`
	Bio      *string `json:"bio" binding:"omitempty,max=255"`
	Location *string `json:"location" binding:"omitempty,max=100"`
	Website  *string `json:"website" binding:"omitempty,url,max=255"`
	Username *string `json:"username" binding:"omitempty,alphanum,min=3,max=30"`
}

type UserResponse struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Verified     bool      `json:"verified"`
	CreatedAt    time.Time `json:"created_at"`
	PasswordHash string    `json:"-"`
	Profile
}

type CreateUserInput struct {
	Email        string
	Username     string
	PasswordHash string
}
