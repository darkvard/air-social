package user

import (
	"time"

	"air-social/internal/domain/media"
	"air-social/internal/domain/user"
)

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
	ObjectKey string              `json:"object_key" binding:"required"`
	Domain    media.UploadDomain  `json:"domain" binding:"required,oneof=users"`
	Feature   media.UploadFeature `json:"feature" binding:"required,oneof=avatar cover"`
}

type BulkConfirmProfileImageRequest struct {
	Files []ConfirmProfileImageRequest `json:"files" binding:"required,min=1,max=10,dive"`
}

type UserDetailResponse struct {
	ID        int64           `json:"id"`
	Email     string          `json:"email"`
	Username  string          `json:"username"`
	Profile   ProfileResponse `json:"profile"`
	Status    StatusResponse  `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Version   int32           `json:"version"`
}

type ProfileResponse struct {
	FullName   string `json:"full_name"`
	Bio        string `json:"bio"`
	Avatar     string `json:"avatar"`
	CoverImage string `json:"cover_image"`
	Location   string `json:"location"`
	Website    string `json:"website"`
}

type StatusResponse struct {
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at"`
}

func NewUserDetailResponse(user *user.User, avatar, cover string) UserDetailResponse {
	return UserDetailResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Profile: ProfileResponse{
			FullName:   user.Profile.FullName,
			Bio:        user.Profile.Bio,
			Avatar:     avatar,
			CoverImage: cover,
			Location:   user.Profile.Location,
			Website:    user.Profile.Website,
		},
		Status: StatusResponse{
			Verified:   user.Status.Verified,
			VerifiedAt: user.Status.VerifiedAt,
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Version:   user.Version,
	}
}
