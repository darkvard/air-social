package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain"
	"air-social/pkg"
)

type MediaService interface {
	GetPresignedURL(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error)
	ConfirmUpload(ctx context.Context, objectName string, userID int64) (string, error)
	DeleteFile(ctx context.Context, fullURL string) error
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

func (s *MediaServiceImpl) GetPresignedURL(ctx context.Context, input domain.PresignedFile) (domain.PresignedFileResponse, error) {
	objectName := s.generateObjectKey(input)
	bucketName := s.cfg.BucketPublic

	uploadURL, err := s.storage.GetPresignedPutURL(ctx, bucketName, objectName, domain.UploadExpiry)
	if err != nil {
		return domain.PresignedFileResponse{}, err
	}

	if err := s.saveUploadSession(ctx, objectName, input.UserID); err != nil {
		return domain.PresignedFileResponse{}, err
	}

	return domain.PresignedFileResponse{
		UploadURL:     uploadURL,
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

	return s.buildPublicURL(objectName), nil
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
	_ = s.cache.Delete(ctx, key)
	return nil
}

func (s *MediaServiceImpl) DeleteFile(ctx context.Context, fullURL string) error {
	if objectName, ok := s.parseObjectNameFromURL(fullURL); ok {
		return s.storage.DeleteFile(ctx, s.cfg.BucketPublic, objectName)
	}
	return nil
}

func (s *MediaServiceImpl) buildPublicURL(objectName string) string {
	return fmt.Sprintf("%s/%s/%s", s.cfg.PublicURL, s.cfg.BucketPublic, objectName)
}

func (s *MediaServiceImpl) parseObjectNameFromURL(fullURL string) (string, bool) {
	prefix := fmt.Sprintf("%s/%s/", s.cfg.PublicURL, s.cfg.BucketPublic)
	if strings.HasPrefix(fullURL, prefix) {
		return strings.TrimPrefix(fullURL, prefix), true
	}
	return "", false
}
