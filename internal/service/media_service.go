package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain"
	"air-social/pkg"
)

// MediaService handles file upload flow:
// 1. Presigned: Client requests a presigned URL (GetPresignedURL).
// 2. Upload: Client uploads file directly to Storage (MinIO/S3) using the URL.
//   - Method: PUT
//   - Body: Binary file data
//   - Header: Host: minio:9000 (Required if running in local Docker env to match signature)
//
// 3. Confirm: Client calls backend to confirm upload (ConfirmUpload). Backend verifies and updates DB.
type MediaService interface {
	GetPresignedURL(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error)
	ConfirmUpload(ctx context.Context, objectName string, userID int64) (string, error)
	DeleteFile(ctx context.Context, fullURL string) error
	GetPublicURL(objectName string) string
}

type MediaServiceImpl struct {
	storage domain.FileStorage
	cache   domain.CacheStorage
	cfg     domain.FileConfig
}

func NewMediaService(storage domain.FileStorage, cache domain.CacheStorage, cfg domain.FileConfig) *MediaServiceImpl {
	return &MediaServiceImpl{
		storage: storage,
		cache:   cache,
		cfg:     cfg,
	}
}

// GetPresignedURL generates a temporary URL for uploading files.
//
// Usage for Client:
//   - Method: PUT
//   - URL: The 'upload_url' returned from this function.
//   - Body: The raw binary content of the file.
//
// Note: Client MUST send 'Host: minio:9000' header so MinIO can validate the signature.
func (s *MediaServiceImpl) GetPresignedURL(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error) {
	objectName := s.generateObjectKey(input)
	bucketName := s.cfg.BucketPublic

	uploadURL, err := s.storage.GetPresignedPutURL(ctx, bucketName, objectName, domain.UploadExpiry)
	if err != nil {
		return domain.PresignedFileResponse{}, err
	}

	// Parse the signed URL returned by MinIO to get the query parameters (signature, etc.)
	u, err := url.Parse(uploadURL)
	if err != nil {
		return domain.PresignedFileResponse{}, fmt.Errorf("failed to parse presigned url: %w", err)
	}

	// Construct the Public URL: PublicDomain + / + ObjectName + ? + QueryParams
	// Example: http://localhost/air-social-public/users/1/avatar.jpg?X-Amz-Signature=...
	finalURL := fmt.Sprintf("%s/%s?%s", s.cfg.PublicURL, objectName, u.RawQuery)

	if err := s.saveUploadSession(ctx, objectName, input.UserID); err != nil {
		return domain.PresignedFileResponse{}, err
	}

	return domain.PresignedFileResponse{
		UploadURL:     finalURL,
		ObjectName:    objectName,
		ExpirySeconds: int64(domain.UploadExpiry.Seconds()),
	}, nil
}

func (s *MediaServiceImpl) generateObjectKey(input domain.PresignedFile) string {
	fileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String(), input.Ext)

	// Flexible folder structure
	// e.g: users/123/avatar/timestamp_uuid.jpg
	// e.g: posts/123/image/timestamp_uuid.jpg
	return fmt.Sprintf("%s/%d/%s/%s", input.Folder, input.UserID, input.Typ, fileName)
}

func (s *MediaServiceImpl) saveUploadSession(ctx context.Context, objectName string, userID int64) error {
	key := domain.GetUploadImageKey(objectName)
	return s.cache.Set(ctx, key, userID, domain.UploadExpiry)
}

func (s *MediaServiceImpl) ConfirmUpload(ctx context.Context, objectName string, userID int64) (string, error) {
	if err := s.verifyUploadSession(ctx, objectName, userID); err != nil {
		return "", err
	}

	exists, err := s.storage.StatFile(ctx, s.cfg.BucketPublic, objectName)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("file not found in storage")
	}

	s.removeUploadSession(ctx, objectName)

	return objectName, nil
}

func (s *MediaServiceImpl) verifyUploadSession(ctx context.Context, objectName string, userID int64) error {
	key := domain.GetUploadImageKey(objectName)
	var cachedUserID int64
	if err := s.cache.Get(ctx, key, &cachedUserID); err != nil {
		return errors.New("upload session expired or invalid")
	}

	if cachedUserID != userID {
		return pkg.ErrUnauthorized
	}
	return nil
}

func (s *MediaServiceImpl) removeUploadSession(ctx context.Context, objectName string) error {
	key := domain.GetUploadImageKey(objectName)
	return s.cache.Delete(ctx, key)
}

func (s *MediaServiceImpl) DeleteFile(ctx context.Context, fullURL string) error {
	objectName := fullURL

	// If input is a full URL, strip the public domain prefix to get the object name
	// Example: http://localhost/public/users/1.jpg -> users/1.jpg
	if strings.HasPrefix(fullURL, "http") {
		prefix := fmt.Sprintf("%s/", s.cfg.PublicURL)
		objectName = strings.TrimPrefix(fullURL, prefix)
	}

	return s.storage.DeleteFile(ctx, s.cfg.BucketPublic, objectName)
}

func (s *MediaServiceImpl) GetPublicURL(objectName string) string {
	if objectName == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.cfg.PublicURL, objectName)
}
