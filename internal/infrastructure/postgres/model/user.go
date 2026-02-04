package model

import (
	"time"

	"air-social/internal/domain"
)

type User struct {
	// Identifier
	ID           int64  `db:"id"`
	Email        string `db:"email"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`

	// Profile
	FullName   string `db:"full_name"`
	Bio        string `db:"bio"`
	Avatar     string `db:"avatar"`
	CoverImage string `db:"cover_image"`
	Location   string `db:"location"`
	Website    string `db:"website"`

	// System info
	Verified   bool       `db:"verified"`
	VerifiedAt *time.Time `db:"verified_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	Version    int        `db:"version"`
}

func (m *User) ToDomain() *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		Profile: domain.Profile{
			FullName:   m.FullName,
			Bio:        m.Bio,
			Avatar:     m.Avatar,
			CoverImage: m.CoverImage,
			Location:   m.Location,
			Website:    m.Website,
		},
		Status: domain.UserStatus{
			Verified:   m.Verified,
			VerifiedAt: m.VerifiedAt,
		},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Version:   m.Version,
	}
}

func FromDomainUser(u *domain.User) *User {
	return &User{
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

func MapToDomainUsers(users []User) []domain.User {
    if users == nil {
        return nil 
    }
    
    result := make([]domain.User, 0, len(users))
    
    for _, user := range users {
        result = append(result, *user.ToDomain())
    }
    
    return result
}