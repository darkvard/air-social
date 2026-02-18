package user

import (
	"time"
)

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
}

type Profile struct {
	FullName   string
	Bio        string
	Avatar     string
	CoverImage string
	Location   string
	Website    string
}

