package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"

	"air-social/internal/domain"
)

// ==============================
// mockUserService
// ==============================
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) GetProfile(ctx context.Context, id int64) (domain.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.UserResponse), args.Error(1)
}

func (m *mockUserService) CreateUser(ctx context.Context, input domain.CreateUserParams) (domain.UserResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(domain.UserResponse), args.Error(1)
}

func (m *mockUserService) UpdateProfile(ctx context.Context, input domain.UpdateProfileParams) (domain.UserResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(domain.UserResponse), args.Error(1)
}

func (m *mockUserService) ChangePassword(ctx context.Context, input domain.ChangePasswordParams) error {
	return m.Called(ctx, input).Error(0)
}

func (m *mockUserService) UpdatePassword(ctx context.Context, email, newPassword string) error {
	return m.Called(ctx, email, newPassword).Error(0)
}

func (m *mockUserService) VerifyEmail(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

func (m *mockUserService) ConfirmImageUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockUserService) ResolveMediaURLs(res *domain.UserResponse) {
	m.Called(res)
}

// ==============================
// mockTokenService
// ==============================
type mockTokenService struct {
	mock.Mock
}

func (m *mockTokenService) CreateSession(ctx context.Context, userID int64, deviceID string) (domain.TokenInfo, error) {
	args := m.Called(ctx, userID, deviceID)
	return args.Get(0).(domain.TokenInfo), args.Error(1)
}

func (m *mockTokenService) Refresh(ctx context.Context, refreshToken string) (domain.TokenInfo, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(domain.TokenInfo), args.Error(1)
}

func (m *mockTokenService) RevokeSingle(ctx context.Context, refreshToken string) error {
	return m.Called(ctx, refreshToken).Error(0)
}

func (m *mockTokenService) Validate(accessToken string) (*jwt.Token, error) {
	args := m.Called(accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *mockTokenService) RevokeDeviceSession(ctx context.Context, userID int64, deviceID string) error {
	return m.Called(ctx, userID, deviceID).Error(0)
}

func (m *mockTokenService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockTokenService) CleanupDatabase(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// ==============================
// mockMediaService
// ==============================
type mockMediaService struct {
	mock.Mock
}

func (m *mockMediaService) GetPresignedURL(ctx context.Context, input domain.PresignedFileParams) (domain.PresignedFileResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(domain.PresignedFileResponse), args.Error(1)

}

func (m *mockMediaService) ConfirmUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockMediaService) DeleteFile(ctx context.Context, fullURL string) error {
	return m.Called(ctx, fullURL).Error(0)
}

func (m *mockMediaService) GetPublicURL(objectName string) string {
	return m.Called(objectName).Get(0).(string)
}

// ==============================
// mockTokenRepository
// ==============================
type mockTokenRepository struct {
	mock.Mock
}

func (m *mockTokenRepository) Create(ctx context.Context, token domain.RefreshToken) error {
	return m.Called(ctx, token).Error(0)
}

func (m *mockTokenRepository) GetByHash(ctx context.Context, hash string) (domain.RefreshToken, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(domain.RefreshToken), args.Error(1)
}

func (m *mockTokenRepository) UpdateRevoked(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockTokenRepository) UpdateRevokedByDevice(ctx context.Context, userID int64, deviceID string) error {
	return m.Called(ctx, userID, deviceID).Error(0)
}

func (m *mockTokenRepository) UpdateRevokedByUser(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockTokenRepository) DeleteExpiredAndRevoked(ctx context.Context, expiredThreshold, revokedThreshold time.Time) error {
	return m.Called(ctx, expiredThreshold, revokedThreshold).Error(0)
}

// ==============================
// mockUserRepository
// ==============================
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)

}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) UpdateProfileImages(ctx context.Context, userID int64, url string, feature domain.UploadFeature) error {
	return m.Called(ctx, userID, url, feature).Error(0)
}

// ==============================
// mockEventQueue
// ==============================
type mockEventQueue struct {
	mock.Mock
}

func (m *mockEventQueue) Publish(ctx context.Context, topic string, payload any) error {
	return m.Called(ctx, topic, payload).Error(0)
}

func (m *mockEventQueue) Close() {
	m.Called()
}

// ==============================
// mockCacheStorage
// ==============================
type mockCacheStorage struct {
	mock.Mock
}

func (m *mockCacheStorage) Get(ctx context.Context, key string, dst any) error {
	args := m.Called(ctx, key, dst)
	return args.Error(0)
}

func (m *mockCacheStorage) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	args := m.Called(ctx, key, val, ttl)
	return args.Error(0)
}

func (m *mockCacheStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockCacheStorage) IsExist(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// ==============================
// mockURLFactory
// ==============================
type mockURLFactory struct {
	mock.Mock
}

func (m *mockURLFactory) ResetPasswordLink(token string) string {
	return m.Called(token).String(0)
}

func (m *mockURLFactory) VerifyEmailLink(token string) string {
	return m.Called(token).String(0)
}

func (m *mockURLFactory) APIRouterPath() string {
	return m.Called().String(0)
}

func (m *mockURLFactory) SwaggerUI() string {
	return m.Called().String(0)
}

func (m *mockURLFactory) MinioConsoleUI() string {
	return m.Called().String(0)

}
func (m *mockURLFactory) RabbitMQDashboardUI() string {
	return m.Called().String(0)
}

func (m *mockURLFactory) FileStorageBaseURL() string {
	return m.Called().String(0)
}
