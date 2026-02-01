package service

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

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
	cfg     domain.FileConfig
}

func NewMediaService(storage domain.FileStorage, cfg domain.FileConfig) *MediaServiceImpl {
	return &MediaServiceImpl{
		storage: storage,
		cfg:     cfg,
	}
}

func (s *MediaServiceImpl) GetPresignedURL(ctx context.Context, input domain.PresignedFileParams) (domain.PresignedFileResponse, error) {
	var empty domain.PresignedFileResponse

	rule, err := s.validateAndGetUploadRule(input)
	if err != nil {
		return empty, err
	}

	loc := domain.StorageLocation{
		Bucket: s.cfg.BucketPublic,
		Key:    s.generateObjectKey(input),
	}
	constraints := domain.UploadConstraints{
		Expiry:      domain.PresignedUploadExpiry,
		ContentType: input.FileType,
		MaxSize:     rule.MaxBytes,
	}

	result, err := s.storage.GetPresignedPostPolicy(ctx, loc, constraints)
	if err != nil {
		return empty, err
	}

	// build public URL using config. MinIO returns internal Docker endpoint not accessible to clients.
	baseURL := strings.TrimSuffix(s.cfg.DomainPublic, "/")
	uploadURL := fmt.Sprintf("%s/%s", baseURL, s.cfg.BucketPublic)

	return domain.PresignedFileResponse{
		UploadURL: uploadURL,
		FormData:  result.FormData,
		ObjectKey: loc.Key,
		PublicURL: s.GetPublicURL(loc.Key),
		ExpireAt:  pkg.TimeNowUTC().Add(domain.PresignedUploadExpiry),
	}, nil
}

func (s *MediaServiceImpl) ConfirmUpload(ctx context.Context, input domain.ConfirmFileParams) (string, error) {
	loc := domain.StorageLocation{
		Bucket: s.cfg.BucketPublic,
		Key:    input.ObjectKey,
	}

	if err := s.validateUploadPath(input); err != nil {
		return "", err
	}

	exists, err := s.storage.StatFile(ctx, loc)
	if err != nil {
		return "", pkg.OrInternalError(err)
	}
	if !exists {
		return "", pkg.ErrNotFound
	}

	return loc.Key, nil
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
	baseURL := strings.TrimSuffix(s.cfg.DomainPublic, "/")
	return fmt.Sprintf("%s/%s/%s", baseURL, s.cfg.BucketPublic, objectKey)
}

// Internal helpers

// Format: {domain}/{entity_id}/{feature}/{timestamp}_{uuid}{ext}
func (s *MediaServiceImpl) generateObjectKey(input domain.PresignedFileParams) string {
	ext := filepath.Ext(input.FileName)
	uid := uuid.New().String()
	timestamp := pkg.TimeNowUTC().Unix()
	fileName := fmt.Sprintf("%d_%s%s", timestamp, uid, ext)
	return fmt.Sprintf("%s/%d/%s/%s", input.Domain, input.EntityID, input.Feature, fileName)
}

func (s *MediaServiceImpl) validateAndGetUploadRule(input domain.PresignedFileParams) (domain.UploadRule, error) {
	var empty domain.UploadRule

	features, ok := domain.FileUploadRules[input.Domain]
	if !ok {
		return empty, fmt.Errorf("%w: domain '%s' is not supported", pkg.ErrBadRequest, input.Domain)
	}

	rule, ok := features[input.Feature]
	if !ok {
		return empty, fmt.Errorf("%w: feature '%s' is not supported for domain '%s'", pkg.ErrBadRequest, input.Feature, input.Domain)
	}

	if input.FileSize > rule.MaxBytes {
		return empty, fmt.Errorf("%w: file size %d bytes exceeds limit of %d bytes", pkg.ErrFileTooLarge, input.FileSize, rule.MaxBytes)
	}

	if !slices.Contains(rule.AllowedTypes, input.FileType) {
		return empty, fmt.Errorf("%w: file type '%s' is not allowed, allowed types: %v", pkg.ErrBadRequest, input.FileType, rule.AllowedTypes)
	}

	return rule, nil
}

func (s *MediaServiceImpl) validateUploadPath(input domain.ConfirmFileParams) error {
	allowedFeatures, ok := domain.ValidDomainFeatures[input.Domain]
	if !ok {
		return fmt.Errorf("%w: domain '%s' not supported", pkg.ErrBadRequest, input.Domain)
	}

	if !slices.Contains(allowedFeatures, input.Feature) {
		return fmt.Errorf("%w: feature '%s' not allowed for domain '%s'", pkg.ErrBadRequest, input.Feature, input.Domain)
	}

	expectedPrefix := fmt.Sprintf("%s/%d/%s/", input.Domain, input.EntityID, input.Feature)

	if !strings.HasPrefix(input.ObjectKey, expectedPrefix) {
		return fmt.Errorf("%w: object key mismatch, expected prefix: %s", pkg.ErrForbidden, expectedPrefix)
	}

	return nil
}
