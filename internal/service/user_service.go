package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

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

	PresignedImageUpload(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error)
	ConfirmImageUpload(ctx context.Context, input domain.ConfirmFile) (string, error)
}

type UserServiceImpl struct {
	repo    domain.UserRepository
	storage domain.FileStorage
	cache   domain.CacheStorage
	cfg     domain.FileConfig
}

func NewUserService(repo domain.UserRepository, storage domain.FileStorage, cache domain.CacheStorage, cfg domain.FileConfig) *UserServiceImpl {
	return &UserServiceImpl{
		repo:    repo,
		storage: storage,
		cache:   cache,
		cfg:     cfg,
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

func (s *UserServiceImpl) PresignedImageUpload(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error) {
	fileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), input.Ext)
	objectName := fmt.Sprintf("users/%d/%s/%s", input.UserID, input.Typ, fileName)
	bucketName := s.getBucketName(true)

	uploadURL, err := s.storage.GetPresignedPutURL(ctx, bucketName, objectName, domain.UploadExpiry)
	if err != nil {
		return domain.PresignedFileResponse{}, err
	}

	if err := s.storeUploadSession(ctx, objectName, input.UserID); err != nil {
		return domain.PresignedFileResponse{}, err
	}

	return domain.PresignedFileResponse{
		UploadURL:     uploadURL,
		ObjectName:    objectName,
		ExpirySeconds: int64(domain.UploadExpiry.Seconds()),
	}, nil
}

func (s *UserServiceImpl) ConfirmImageUpload(ctx context.Context, input domain.ConfirmFile) (string, error) {
	bucketName := s.getBucketName(true)

	if err := s.verifyUploadSession(ctx, input.ObjectName, input.UserID); err != nil {
		return "", err
	}

	exists, err := s.storage.StatFile(ctx, bucketName, input.ObjectName)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("file not found or upload failed")
	}

	fullPublicURL := s.buildPublicURL(bucketName, input.ObjectName)
	if err = s.repo.UpdateProfileImages(ctx, input.UserID, fullPublicURL, input.Typ); err != nil {
		return "", err
	}
	return fullPublicURL, nil
}

func (s *UserServiceImpl) getBucketName(isPublic bool) string {
	if isPublic {
		return s.cfg.BucketPublic
	}
	return s.cfg.BucketPrivate
}

func (s *UserServiceImpl) buildPublicURL(bucket, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, bucket, objectName)
}

func (s *UserServiceImpl) storeUploadSession(ctx context.Context, objectName string, userID int64) error {
	key := domain.GetUploadImageKey(objectName)
	return s.cache.Set(ctx, key, userID, domain.UploadExpiry)
}

func (s *UserServiceImpl) verifyUploadSession(ctx context.Context, objectName string, userID int64) error {
	key := domain.GetUploadImageKey(objectName)
	var cachedUserID int64
	if err := s.cache.Get(ctx, key, &cachedUserID); err != nil {
		return errors.New("upload session expired or invalid")
	}

	if cachedUserID != userID {
		return pkg.ErrUnauthorized
	}
	_ = s.cache.Delete(ctx, key)
	return nil
}
