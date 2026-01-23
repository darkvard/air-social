package service

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"air-social/internal/domain"
	"air-social/pkg"
)

type MediaService interface {
	GetPresignedURL(ctx context.Context, input domain.PresignedFileParams) (domain.PresignedFileResponse, error)
	ConfirmUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error)
	DeleteFile(ctx context.Context, objectKey string) error
	GetPublicURL(objectKey string) string
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

func (s *MediaServiceImpl) GetPresignedURL(ctx context.Context, input domain.PresignedFileParams) (domain.PresignedFileResponse, error) {
	rule, err := s.getValidationRules(input.Domain, input.Feature)
	if err != nil {
		return domain.PresignedFileResponse{}, err
	}

	if err := s.validateRequest(input, rule); err != nil {
		return domain.PresignedFileResponse{}, err
	}

	loc := domain.StorageLocation{
		Bucket: s.cfg.BucketPublic,
		Key:    s.generateObjectKey(input),
	}
	constraints := domain.UploadConstraints{
		Expiry:      domain.UploadExpiry,
		ContentType: input.FileType,
		MaxSize:     rule.MaxBytes,
	}

	result, err := s.storage.GetPresignedPostPolicy(ctx, loc, constraints)
	if err != nil {
		return domain.PresignedFileResponse{}, err
	}
	if err := s.saveUploadSession(ctx, loc.Key, input.UserID); err != nil {
		return domain.PresignedFileResponse{}, err
	}

	return domain.PresignedFileResponse{
		UploadURL:     result.UploadURL,
		FormData:      result.FormData,
		ObjectKey:     loc.Key,
		PublicURL:     s.GetPublicURL(loc.Key),
		ExpirySeconds: int64(domain.UploadExpiry.Seconds()),
	}, nil
}

func (s *MediaServiceImpl) getValidationRules(d domain.UploadDomain, f domain.UploadFeature) (domain.UploadRule, error) {
	switch d {
	case domain.DomainUser:
		if f == domain.FeatureAvatar || f == domain.FeatureCover {
			return domain.UploadRule{MaxBytes: domain.Limit5MB, AllowedTypes: domain.ImageAllowedTypes}, nil
		}

	case domain.DomainPost:
		if f == domain.FeatureFeedImage {
			return domain.UploadRule{MaxBytes: domain.Limit10MB, AllowedTypes: domain.ImageAllowedTypes}, nil
		}
		if f == domain.FeatureFeedVideo {
			return domain.UploadRule{MaxBytes: domain.Limit100MB, AllowedTypes: domain.VideoAllowedTypes}, nil
		}

	case domain.DomainMessage:
		if f == domain.FeatureVoiceChat {
			return domain.UploadRule{MaxBytes: domain.Limit10MB, AllowedTypes: domain.AudioAllowedTypes}, nil
		}
		if f == domain.FeatureFeedImage {
			return domain.UploadRule{MaxBytes: domain.Limit10MB, AllowedTypes: domain.ImageAllowedTypes}, nil
		}
	}

	return domain.UploadRule{}, pkg.ErrFileUnsupported
}

func (s *MediaServiceImpl) validateRequest(input domain.PresignedFileParams, rule domain.UploadRule) error {
	if input.FileSize > rule.MaxBytes {
		return fmt.Errorf("%w (limit: %d bytes)", pkg.ErrFileTooLarge, rule.MaxBytes)
	}

	isValidType := slices.Contains(rule.AllowedTypes, input.FileType)

	if !isValidType {
		return fmt.Errorf("%w (allowed: %s)", pkg.ErrFileTypeInvalid, strings.Join(rule.AllowedTypes, ", "))
	}

	return nil
}

// Format: {domain}/{entity_id}/{feature}/{timestamp}_{uuid}{ext}
func (s *MediaServiceImpl) generateObjectKey(input domain.PresignedFileParams) string {
	ext := filepath.Ext(input.FileName)
	uid := uuid.New().String()
	timestamp := time.Now().Unix()
	fileName := fmt.Sprintf("%d_%s%s", timestamp, uid, ext)
	return fmt.Sprintf("%s/%d/%s/%s", input.Domain, input.UserID, input.Feature, fileName)
}

func (s *MediaServiceImpl) saveUploadSession(ctx context.Context, objectName string, userID int64) error {
	key := domain.GetUploadImageKey(objectName)
	return s.cache.Set(ctx, key, userID, domain.UploadExpiry)
}

func (s *MediaServiceImpl) ConfirmUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	loc := domain.StorageLocation{
		Bucket: s.cfg.BucketPublic,
		Key:    input.ObjectKey,
	}

	if err := s.verifyUploadSession(ctx, loc.Key, input.UserID); err != nil {
		return "", err
	}

	exists, err := s.storage.StatFile(ctx, loc)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", pkg.ErrNotFound
	}

	s.removeUploadSession(ctx, loc.Key)

	return loc.Key, nil
}

func (s *MediaServiceImpl) verifyUploadSession(ctx context.Context, objectName string, userID int64) error {
	var cachedUserID int64
	if err := s.cache.Get(ctx, domain.GetUploadImageKey(objectName), &cachedUserID); err != nil {
		return pkg.ErrSessionExpired
	}
	if cachedUserID != userID {
		return pkg.ErrUnauthorized
	}
	return nil
}

func (s *MediaServiceImpl) removeUploadSession(ctx context.Context, objectName string) error {
	return s.cache.Delete(ctx, domain.GetUploadImageKey(objectName))
}

func (s *MediaServiceImpl) DeleteFile(ctx context.Context, objectKey string) error {
	loc := domain.StorageLocation{
		Bucket: s.cfg.BucketPublic,
		Key:    objectKey,
	}

	return s.storage.DeleteFile(ctx, loc)
}

func (s *MediaServiceImpl) GetPublicURL(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	baseURL := strings.TrimSuffix(s.cfg.PublicPathPrefix, "/")
	return fmt.Sprintf("%s/%s", baseURL, objectKey)
}
