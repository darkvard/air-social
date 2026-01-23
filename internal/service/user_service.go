package service

import (
	"context"
	"time"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserService interface {
	CreateUser(ctx context.Context, req domain.CreateUserParams) (domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (domain.UserResponse, error)
	GetByID(ctx context.Context, id int64) (domain.UserResponse, error)
	VerifyEmail(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, password string) error
	ChangePassword(ctx context.Context, userID int64, req domain.ChangePasswordRequest) error
	UpdateProfile(ctx context.Context, userID int64, req domain.UpdateProfileRequest) (domain.UserResponse, error)
	ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error)
}

type UserServiceImpl struct {
	repo  domain.UserRepository
	media MediaService
}

func NewUserService(repo domain.UserRepository, media MediaService) *UserServiceImpl {
	return &UserServiceImpl{
		repo:  repo,
		media: media,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, in domain.CreateUserParams) (domain.UserResponse, error) {
	empty := domain.UserResponse{}
	if existing, _ := s.repo.GetByEmail(ctx, in.Email); existing != nil {
		return empty, pkg.ErrAlreadyExists
	}

	u := &domain.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: in.PasswordHash,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return empty, err
	}

	return s.mapToResponse(u), nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (domain.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, email string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	user.Verified = true
	user.VerifiedAt = &now
	return s.repo.Update(ctx, user)
}

func (s *UserServiceImpl) UpdatePassword(ctx context.Context, email, pwd string) error {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.PasswordHash = pwd
	return s.repo.Update(ctx, user)
}

func (s *UserServiceImpl) ChangePassword(ctx context.Context, userID int64, req domain.ChangePasswordRequest) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if req.NewPassword == req.CurrentPassword {
		return pkg.ErrSamePassword
	}

	if !verifyPassword(req.CurrentPassword, user.PasswordHash) {
		return pkg.ErrInvalidCredentials
	}

	hashedPwd, err := hashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPwd
	return s.repo.Update(ctx, user)
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID int64, req domain.UpdateProfileRequest) (domain.UserResponse, error) {
	empty := domain.UserResponse{}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return empty, err
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Location != nil {
		user.Location = *req.Location
	}
	if req.Website != nil {
		user.Website = *req.Website
	}
	if req.Username != nil {
		user.Username = *req.Username
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return empty, err
	}

	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	if input.Feature != domain.FeatureAvatar && input.Feature != domain.FeatureCover {
		return "", pkg.ErrInvalidData
	}

	objectKey, err := s.media.ConfirmUpload(ctx, input)
	if err != nil {
		return "", err
	}

	if err = s.repo.UpdateProfileImages(ctx, input.UserID, objectKey, input.Feature); err != nil {
		return "", err
	}

	return s.media.GetPublicURL(objectKey), nil
}

func (s *UserServiceImpl) mapToResponse(u *domain.User) domain.UserResponse {
	res := u.ToResponse()
	res.Avatar = s.media.GetPublicURL(res.Avatar)
	res.CoverImage = s.media.GetPublicURL(res.CoverImage)
	return res
}
