package service

// import (
// 	"context"
// 	"fmt"
// 	"path/filepath"
// 	"slices"
// 	"strings"
//
//

// 	"github.com/google/uuid"

// 	"air-social/internal/config"
// 	"air-social/internal/domain"
// 	"air-social/pkg"
// )

// type MediaService interface {
// 	GetPresignedURL(ctx context.Context, input []domain.PresignedFileParams) ([]domain.PresignedFile, error)
// 	ConfirmUpload(ctx context.Context, input []domain.ConfirmFileParams) ([]string, error)
// 	DeleteFile(ctx context.Context, objectKeys []string) error
// 	VerifyMedia(ctx context.Context, objectKeys []string) error
// }

// type MediaVerifier interface {
// 	VerifyMedia(ctx context.Context, objectKeys []string) error
// }

// type MediaUploadConfirmer interface {
// 	ConfirmUpload(ctx context.Context, input []domain.ConfirmFileParams) ([]string, error)
// }

// type MediaServiceImpl struct {
// 	storage    domain.FileStorage
// 	bucket     domain.FileBucket
// 	urlFactory domain.URLFactory
// }

// func NewMediaService(storage domain.FileStorage, cfg config.MinioStorageConfig, urlFactory domain.URLFactory) *MediaServiceImpl {
// 	return &MediaServiceImpl{
// 		storage: storage,
// 		bucket: domain.FileBucket{
// 			Public:  cfg.BucketPublic,
// 			Private: cfg.BucketPrivate,
// 		},
// 		urlFactory: urlFactory,
// 	}
// }

// func (s *MediaServiceImpl) GetPresignedURL(ctx context.Context, input []domain.PresignedFileParams) ([]domain.PresignedFile, error) {
// 	results := make([]domain.PresignedFile, 0, len(input))

// 	for _, item := range input {
// 		rule, err := s.validateAndGetUploadRule(item)
// 		if err != nil {
// 			return nil, err
// 		}

// 		loc := domain.StorageLocation{
// 			Bucket: s.bucket.Public,
// 			Key:    s.generateObjectKey(item),
// 		}
// 		constraints := domain.UploadConstraints{
// 			Expiry:      domain.PresignedUploadExpiry,
// 			ContentType: item.FileType,
// 			MaxSize:     rule.MaxBytes,
// 		}

// 		result, err := s.storage.GetPresignedPostPolicy(ctx, loc, constraints)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// build public URL using config. MinIO returns internal Docker endpoint not accessible to clients.
// 		baseURL := strings.TrimSuffix(s.urlFactory.BaseURL(), "/")
// 		uploadURL := fmt.Sprintf("%s/%s", baseURL, s.bucket.Public)

// 		results = append(results, domain.PresignedFile{
// 			FileName:  item.FileName,
// 			UploadURL: uploadURL,
// 			FormData:  result.FormData,
// 			ObjectKey: loc.Key,
// 			PublicURL: s.urlFactory.PublicFileURL(loc.Key),
// 			ExpireAt:  pkg.TimeNowUTC().Add(domain.PresignedUploadExpiry),
// 		})
// 	}

// 	return results, nil
// }

// func (s *MediaServiceImpl) ConfirmUpload(ctx context.Context, input []domain.ConfirmFileParams) ([]string, error) {
// 	results := make([]string, 0, len(input))

// 	for _, item := range input {
// 		if err := s.validateUploadPath(item); err != nil {
// 			return nil, err
// 		}

// 		loc := domain.StorageLocation{
// 			Bucket: s.bucket.Public,
// 			Key:    item.ObjectKey,
// 		}

// 		exists, err := s.storage.StatFile(ctx, loc)
// 		if err != nil {
// 			return nil, pkg.OrInternalError(err)
// 		}
// 		if !exists {
// 			return nil, pkg.ErrNotFound
// 		}
// 		results = append(results, loc.Key)
// 	}

// 	return results, nil
// }

// func (s *MediaServiceImpl) DeleteFile(ctx context.Context, objectKeys []string) error {
// 	for _, key := range objectKeys {
// 		loc := domain.StorageLocation{
// 			Bucket: s.bucket.Public,
// 			Key:    key,
// 		}
// 		if err := s.storage.DeleteFile(ctx, loc); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (s *MediaServiceImpl) VerifyMedia(ctx context.Context, objectKeys []string) error {
// 	for _, key := range objectKeys {
// 		loc := domain.StorageLocation{
// 			Bucket: s.bucket.Public,
// 			Key:    key,
// 		}

// 		exists, err := s.storage.StatFile(ctx, loc)
// 		if err != nil {
// 			return err
// 		}
// 		if !exists {
// 			return pkg.ErrNotFound
// 		}
// 	}
// 	return nil
// }

// // Internal helpers

// // Format: {domain}/{entity_id}/{feature}/{timestamp}_{uuid}{ext}
// func (s *MediaServiceImpl) generateObjectKey(input domain.PresignedFileParams) string {
// 	ext := filepath.Ext(input.FileName)
// 	uid := uuid.New().String()
// 	timestamp := pkg.TimeNowUTC().Unix()
// 	fileName := fmt.Sprintf("%d_%s%s", timestamp, uid, ext)
// 	return fmt.Sprintf("%s/%d/%s/%s", input.Domain, input.EntityID, input.Feature, fileName)
// }

// func (s *MediaServiceImpl) validateAndGetUploadRule(input domain.PresignedFileParams) (domain.UploadRule, error) {
// 	var empty domain.UploadRule

// 	features, ok := domain.FileUploadRules[input.Domain]
// 	if !ok {
// 		return empty, fmt.Errorf("%w: domain '%s' is not supported", pkg.ErrBadRequest, input.Domain)
// 	}

// 	rule, ok := features[input.Feature]
// 	if !ok {
// 		return empty, fmt.Errorf("%w: feature '%s' is not supported for domain '%s'", pkg.ErrBadRequest, input.Feature, input.Domain)
// 	}

// 	if input.FileSize > rule.MaxBytes {
// 		return empty, fmt.Errorf("%w: file size %d bytes exceeds limit of %d bytes", pkg.ErrFileTooLarge, input.FileSize, rule.MaxBytes)
// 	}

// 	if !slices.Contains(rule.AllowedTypes, input.FileType) {
// 		return empty, fmt.Errorf("%w: file type '%s' is not allowed, allowed types: %v", pkg.ErrBadRequest, input.FileType, rule.AllowedTypes)
// 	}

// 	return rule, nil
// }

// func (s *MediaServiceImpl) validateUploadPath(input domain.ConfirmFileParams) error {
// 	allowedFeatures, ok := domain.ValidDomainFeatures[input.Domain]
// 	if !ok {
// 		return fmt.Errorf("%w: domain '%s' not supported", pkg.ErrBadRequest, input.Domain)
// 	}

// 	if !slices.Contains(allowedFeatures, input.Feature) {
// 		return fmt.Errorf("%w: feature '%s' not allowed for domain '%s'", pkg.ErrBadRequest, input.Feature, input.Domain)
// 	}

// 	expectedPrefix := fmt.Sprintf("%s/%d/%s/", input.Domain, input.EntityID, input.Feature)

// 	if !strings.HasPrefix(input.ObjectKey, expectedPrefix) {
// 		return fmt.Errorf("%w: object key mismatch, expected prefix: %s", pkg.ErrForbidden, expectedPrefix)
// 	}

// 	return nil
// }
