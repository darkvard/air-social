package user

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Profile      any       `db:"profile" json:"profile"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	Version      int       `db:"version" json:"version"`
}


type UpdateRequest struct {
	Email    *string `json:"email,omitempty"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Profile  any     `json:"profile,omitempty"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Profile   any       `json:"profile,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}


type CreateUserInput struct {
    Email        string
    Username     string
    PasswordHash string
}
