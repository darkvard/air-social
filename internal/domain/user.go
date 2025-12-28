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
	ID           int64      `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Profile      any        `db:"profile" json:"profile"`
	Verified     bool       `db:"verified" json:"verified"`
	VerifiedAt   *time.Time `db:"verified_at" json:"verified_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	Version      int        `db:"version" json:"version"`
}

type UpdateRequest struct {
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Profile  any     `json:"profile,omitempty"`
}

type UserResponse struct {
	ID         int64      `json:"id"`
	Email      string     `json:"email"`
	Username   string     `json:"username"`
	Profile    any        `json:"profile,omitempty"`
	Verified   bool       `json:"verified"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateUserInput struct {
	Email        string
	Username     string
	PasswordHash string
}
