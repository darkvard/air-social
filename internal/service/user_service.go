package service

import (
	"context"
	"mime/multipart"
	"time"

	"air-social/internal/domain"
	"air-social/pkg"
)

// todo: ko dung con tro cho dto
type UserService interface {
	CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error)
	GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error)
	GetByID(ctx context.Context, id int64) (*domain.UserResponse, error)
	VerifyEmail(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, pwd string) error
	ChangePassword(ctx context.Context, userID int64, req *domain.ChangePasswordRequest) error
	UpdateProfile(ctx context.Context, userID int64, req *domain.UpdateProfileRequest) (*domain.UserResponse, error)
	UpdateAvatar(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (string, error)
}

type UserServiceImpl struct {
	repo    domain.UserRepository
	storage domain.FileStorage
}

func NewUserService(repo domain.UserRepository, storage domain.FileStorage) *UserServiceImpl {
	return &UserServiceImpl{
		repo:    repo,
		storage: storage,
	}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, in *domain.CreateUserInput) (*domain.UserResponse, error) {
	if existing, _ := s.repo.GetByEmail(ctx, in.Email); existing != nil {
		return nil, pkg.ErrAlreadyExists
	}

	u := &domain.User{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: in.PasswordHash,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u.ToResponse(), nil
}

func (s *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (*domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
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

func (s *UserServiceImpl) ChangePassword(ctx context.Context, userID int64, req *domain.ChangePasswordRequest) error {
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

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, userID int64, req *domain.UpdateProfileRequest) (*domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return user.ToResponse(), nil
}

func (s *UserServiceImpl) UpdateAvatar(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	avatarURL, err := s.storage.UploadFile(ctx, file, fileHeader, "avatars")
	if err != nil {
		return "", err
	}

	return avatarURL, nil
}
