package service

import (
	"context"
	"time"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserService interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetProfile(ctx context.Context, id int64) (domain.UserResponse, error)

	CreateUser(ctx context.Context, input domain.CreateUserParams) (domain.UserResponse, error)
	UpdateProfile(ctx context.Context, input domain.UpdateProfileParams) (domain.UserResponse, error)
	ChangePassword(ctx context.Context, input domain.ChangePasswordParams) error
	UpdatePassword(ctx context.Context, email, passwordHashed string) error
	VerifyEmail(ctx context.Context, email string) error

	ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error)
	ResolveMediaURLs(res *domain.UserResponse)
}

type UserServiceImpl struct {
	userRepo domain.UserRepository
	mediaSvc MediaService
}

func NewUserService(userRepo domain.UserRepository, mediaSvc MediaService) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		mediaSvc: mediaSvc,
	}
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return user, nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	return user, nil
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, id int64) (domain.UserResponse, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) ResolveMediaURLs(res *domain.UserResponse) {
	if res == nil {
		return
	}
	if res.Avatar != "" {
		res.Avatar = s.mediaSvc.GetPublicURL(res.Avatar)
	}
	if res.CoverImage != "" {
		res.CoverImage = s.mediaSvc.GetPublicURL(res.CoverImage)
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, input domain.CreateUserParams) (domain.UserResponse, error) {
	var empty domain.UserResponse

	exists, err := s.GetByEmail(ctx, input.Email)
	if err := pkg.SkipError(err, pkg.ErrNotFound); err != nil {
		return empty, err
	}
	if exists != nil {
		return empty, pkg.ErrAlreadyExists
	}

	user := &domain.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: input.PasswordHashed,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return empty, pkg.OrInternalError(err)
	}

	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, input domain.UpdateProfileParams) (domain.UserResponse, error) {
	var empty domain.UserResponse

	user, err := s.GetByID(ctx, input.UserID)
	if err != nil {
		return empty, err
	}

	if input.FullName != nil {
		user.FullName = *input.FullName
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.Location != nil {
		user.Location = *input.Location
	}
	if input.Website != nil {
		user.Website = *input.Website
	}
	if input.Username != nil {
		user.Username = *input.Username
	}

	if err := s.updateUser(ctx, user); err != nil {
		return empty, err
	}
	return s.mapToResponse(user), nil
}

func (s *UserServiceImpl) ChangePassword(ctx context.Context, input domain.ChangePasswordParams) error {
	user, err := s.GetByID(ctx, input.UserID)
	if err != nil {
		return err
	}

	if input.NewPassword == input.CurrentPassword {
		return pkg.ErrSamePassword
	}
	if !verifyPassword(input.CurrentPassword, user.PasswordHash) {
		return pkg.ErrInvalidCredentials
	}

	hashedPwd, err := hashPassword(input.NewPassword)
	if err != nil {
		return pkg.OrInternalError(err)
	}
	user.PasswordHash = hashedPwd

	return s.updateUser(ctx, user)
}

func (s *UserServiceImpl) UpdatePassword(ctx context.Context, email, passwordHashed string) error {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	user.PasswordHash = passwordHashed
	return s.updateUser(ctx, user)
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, email string) error {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	user.Verified = true
	user.VerifiedAt = &(now)

	return s.updateUser(ctx, user)
}

func (s *UserServiceImpl) ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	if input.Feature != domain.FeatureAvatar && input.Feature != domain.FeatureCover {
		return "", pkg.ErrInvalidData
	}

	objectKey, err := s.mediaSvc.ConfirmUpload(ctx, input)
	if err != nil {
		return "", pkg.OrInternalError(err, pkg.ErrBadRequest, pkg.ErrForbidden, pkg.ErrNotFound)
	}

	if err = s.userRepo.UpdateProfileImages(ctx, input.UserID, objectKey, input.Feature); err != nil {
		return "", pkg.OrInternalError(err)
	}

	return s.mediaSvc.GetPublicURL(objectKey), nil
}

// Internal helpers

func (s *UserServiceImpl) mapToResponse(user *domain.User) domain.UserResponse {
	res := user.ToResponse()
	s.ResolveMediaURLs(&res)
	return res
}

func (s *UserServiceImpl) updateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}
