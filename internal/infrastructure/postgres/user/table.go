package user

import (
	"time"

	"air-social/internal/domain/user"
)

type Table struct {
	ID           int64      `db:"id"`
	Email        string     `db:"email"`
	Username     string     `db:"username"`
	PasswordHash string     `db:"password_hash"`
	FullName     string     `db:"full_name"`
	Bio          string     `db:"bio"`
	Avatar       string     `db:"avatar"`
	CoverImage   string     `db:"cover_image"`
	Location     string     `db:"location"`
	Website      string     `db:"website"`
	Verified     bool       `db:"verified"`
	VerifiedAt   *time.Time `db:"verified_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	Version      int32      `db:"version"`
}

func (m *Table) ToDomain() *user.User {
	return &user.User{
		ID:           m.ID,
		Email:        m.Email,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		Profile: user.Profile{
			FullName:   m.FullName,
			Bio:        m.Bio,
			Avatar:     m.Avatar,
			CoverImage: m.CoverImage,
			Location:   m.Location,
			Website:    m.Website,
		},
		Status: user.Status{
			Verified:   m.Verified,
			VerifiedAt: m.VerifiedAt,
		},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Version:   m.Version,
	}
}

func FromDomain(u *user.User) *Table {
	return &Table{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		FullName:     u.Profile.FullName,
		Bio:          u.Profile.Bio,
		Avatar:       u.Profile.Avatar,
		CoverImage:   u.Profile.CoverImage,
		Location:     u.Profile.Location,
		Website:      u.Profile.Website,
		Verified:     u.Status.Verified,
		VerifiedAt:   u.Status.VerifiedAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Version:      u.Version,
	}
}
