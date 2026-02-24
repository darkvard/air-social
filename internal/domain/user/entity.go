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
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      int32
}

type Status struct {
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
