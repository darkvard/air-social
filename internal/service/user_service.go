package service

import (
	"context"

	"air-social/internal/domain"
	"air-social/pkg"
)

type UserService interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetProfile(ctx context.Context, id int64) (*domain.User, error)

	CreateUser(ctx context.Context, input domain.CreateUserParams) (*domain.User, error)
	UpdateProfile(ctx context.Context, input domain.UpdateProfileParams) (*domain.User, error)
	ChangePassword(ctx context.Context, input domain.ChangePasswordParams) error
	UpdatePassword(ctx context.Context, email, passwordHashed string) error
	VerifyEmail(ctx context.Context, email string) error

	ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error)
	FormatPublicURL(key string) string
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

func (s *UserServiceImpl) GetProfile(ctx context.Context, id int64) (*domain.User, error) {
	return s.GetByID(ctx, id)
}

func (s *UserServiceImpl) FormatPublicURL(key string) string {
	return s.mediaSvc.FormatPublicURL(key)
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, input domain.CreateUserParams) (*domain.User, error) {

	user := &domain.User{
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: input.PasswordHashed,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrAlreadyExists)
	}

	return user, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, input domain.UpdateProfileParams) (*domain.User, error) {
	var empty *domain.User

	user, err := s.GetByID(ctx, input.UserID)
	if err != nil {
		return empty, err
	}

	if input.FullName != nil {
		user.Profile.FullName = *input.FullName
	}
	if input.Bio != nil {
		user.Profile.Bio = *input.Bio
	}
	if input.Location != nil {
		user.Profile.Location = *input.Location
	}
	if input.Website != nil {
		user.Profile.Website = *input.Website
	}
	if input.Username != nil {
		user.Username = *input.Username
	}

	if err := s.updateUser(ctx, user); err != nil {
		return empty, err
	}
	return user, nil
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

	now := pkg.TimeNowUTC()
	user.Status.Verified = true
	user.Status.VerifiedAt = &(now)

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

	if err = s.userRepo.UpdateProfileImages(ctx, input.EntityID, objectKey, input.Feature); err != nil {
		return "", pkg.OrInternalError(err)
	}

	return s.mediaSvc.FormatPublicURL(objectKey), nil
}

// Internal helpers

func (s *UserServiceImpl) updateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}
