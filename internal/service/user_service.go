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
	GetSummary(ctx context.Context, id int64) (*domain.UserSummary, error)
	CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error)
	UpdateProfile(ctx context.Context, params domain.UpdateProfileParams) (*domain.User, error)
	ChangePassword(ctx context.Context, params domain.ChangePasswordParams) error
	UpdatePassword(ctx context.Context, email, passwordHashed string) error
	VerifyEmail(ctx context.Context, email string) error
	ConfirmImageUpload(ctx context.Context, input []domain.ConfirmFileParams) ([]domain.ConfirmFileResult, error)
}

type UserServiceImpl struct {
	userRepo   domain.UserRepository
	cache      domain.CacheStorage
	urlFactory domain.URLFactory
	mediaSvc   MediaService
}

func NewUserService(userRepo domain.UserRepository, mediaSvc MediaService, cache domain.CacheStorage, urlFactory domain.URLFactory) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:   userRepo,
		cache:      cache,
		urlFactory: urlFactory,
		mediaSvc:   mediaSvc,
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

func (s *UserServiceImpl) GetSummary(ctx context.Context, id int64) (*domain.UserSummary, error) {
	key := domain.GetUserSummaryKey(id)

	var cached domain.UserSummary
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return &cached, nil
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	info := s.toUserSummary(user)
	_ = s.cache.Set(ctx, key, info, 12*time.Hour)

	return info, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error) {
	user := &domain.User{
		Email:        params.Email,
		Username:     params.Username,
		PasswordHash: params.PasswordHashed,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrAlreadyExists)
	}

	return user, nil
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, params domain.UpdateProfileParams) (*domain.User, error) {
	var empty *domain.User

	user, err := s.GetByID(ctx, params.UserID)
	if err != nil {
		return empty, err
	}

	if params.FullName != nil {
		user.Profile.FullName = *params.FullName
	}
	if params.Bio != nil {
		user.Profile.Bio = *params.Bio
	}
	if params.Location != nil {
		user.Profile.Location = *params.Location
	}
	if params.Website != nil {
		user.Profile.Website = *params.Website
	}
	if params.Username != nil {
		user.Username = *params.Username
	}

	if err := s.updateUser(ctx, user); err != nil {
		return empty, err
	}

	_ = s.cache.Delete(ctx, domain.GetUserSummaryKey(user.ID))
	return user, nil
}

func (s *UserServiceImpl) ChangePassword(ctx context.Context, params domain.ChangePasswordParams) error {
	user, err := s.GetByID(ctx, params.UserID)
	if err != nil {
		return err
	}

	if params.NewPassword == params.CurrentPassword {
		return pkg.ErrSamePassword
	}
	if !verifyPassword(params.CurrentPassword, user.PasswordHash) {
		return pkg.ErrInvalidCredentials
	}

	hashedPwd, err := hashPassword(params.NewPassword)
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

	_ = s.cache.Delete(ctx, domain.GetUserSummaryKey(user.ID))
	return s.updateUser(ctx, user)
}

func (s *UserServiceImpl) ConfirmImageUpload(ctx context.Context, params []domain.ConfirmFileParams) ([]domain.ConfirmFileResult, error) {
	for _, input := range params {
		if input.Feature != domain.FeatureAvatar && input.Feature != domain.FeatureCover {
			return nil, pkg.ErrInvalidData
		}
	}

	keys, err := s.mediaSvc.ConfirmUpload(ctx, params)
	if err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrBadRequest, pkg.ErrForbidden, pkg.ErrNotFound)
	}

	responses := make([]domain.ConfirmFileResult, len(params))
	for i, input := range params {
		if err = s.userRepo.UpdateProfileImages(ctx, input.EntityID, keys[i], input.Feature); err != nil {
			return nil, pkg.OrInternalError(err)
		}

		responses[i] = domain.ConfirmFileResult{
			Domain:  input.Domain,
			Feature: input.Feature,
			URL:     s.urlFactory.PublicFileURL(keys[i]),
		}
	}

	if len(params) > 0 {
		_ = s.cache.Delete(ctx, domain.GetUserSummaryKey(params[0].EntityID))
	}
	return responses, nil
}

// Internal helpers

func (s *UserServiceImpl) updateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (s *UserServiceImpl) toUserSummary(user *domain.User) *domain.UserSummary {
	return &domain.UserSummary{
		ID:         user.ID,
		FullName:   user.Profile.FullName,
		Avatar:     s.urlFactory.PublicFileURL(user.Profile.Avatar),
		CoverImage: s.urlFactory.PublicFileURL(user.Profile.CoverImage),
		Verified:   user.Status.Verified,
	}
}
