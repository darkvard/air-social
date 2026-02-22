package user

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, user *User) error

	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)

	UpdateProfile(ctx context.Context, user *User) error
	UpdateAvatar(ctx context.Context, id int64, url string) error
	UpdateCover(ctx context.Context, id int64, url string) error
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateVerified(ctx context.Context, id int64, status bool) error
}
