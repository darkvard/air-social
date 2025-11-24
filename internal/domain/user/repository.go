package user

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
