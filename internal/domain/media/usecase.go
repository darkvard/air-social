package media

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type UseCase interface {
	GenerateKey(params PresignParams) string
	GetPresignedURL(ctx context.Context, params []PresignParams) ([]PresignedResult, error)
	ConfirmUpload(ctx context.Context, params []ConfirmParams) ([]string, error)
	VerifyMedia(ctx context.Context, keys []string) error
	DeleteMedia(ctx context.Context, keys []string) error
}

type Deps struct {
	Bucket  Bucket
	Storage Storage
	Link    common.AppLinkManager
}

type usecase struct {
	deps Deps
}

func NewUseCase(d Deps) *usecase {
	return &usecase{deps: d}
}

// Format: {domain}/{entity_id}/{feature}/{ulid}{ext}
func (u *usecase) GenerateKey(params PresignParams) string {
	ext := filepath.Ext(params.FileName)
	fileName := fmt.Sprintf("%s%s", pkg.NewULID(), ext)
	return fmt.Sprintf("%s/%d/%s/%s", params.Domain, params.EntityID, params.Feature, fileName)
}

func (u *usecase) GetPresignedURL(ctx context.Context, params []PresignParams) ([]PresignedResult, error) {
	results := make([]PresignedResult, 0, len(params))

	for _, item := range params {
		rule, err := validatePresignParams(item)
		if err != nil {
			return nil, err
		}

		location := Location{
			Bucket: u.deps.Bucket.Public,
			Key:    u.GenerateKey(item),
		}
		constraints := Constraints{
			Expiry:      PresignedUploadExpiry,
			ContentType: item.FileType,
			MaxSize:     rule.MaxBytes,
		}

		result, err := u.deps.Storage.PresignUpload(ctx, location, constraints)
		if err != nil {
			return nil, err
		}

		results = append(results, PresignedResult{
			FileName:  item.FileName,
			UploadURL: u.getUploadUrl(),
			FormData:  result.Fields,
			ObjectKey: location.Key,
			PublicURL: u.deps.Link.LinkProvider.PublicFile(location.Key),
			ExpireAt:  pkg.TimeNowUTC().Add(PresignedUploadExpiry),
		})
	}

	return results, nil
}

func (u *usecase) ConfirmUpload(ctx context.Context, params []ConfirmParams) ([]string, error) {
	results := make([]string, 0, len(params))

	for _, item := range params {
		if err := validateConfirmParams(item); err != nil {
			return nil, err
		}

		if err := u.ensureObjectExists(ctx, item.ObjectKey); err != nil {
			return nil, err
		}

		results = append(results, item.ObjectKey)
	}

	return results, nil
}

func (u *usecase) VerifyMedia(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := u.ensureObjectExists(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (u *usecase) DeleteMedia(ctx context.Context, keys []string) error {
	for _, key := range keys {
		location := Location{
			Bucket: u.deps.Bucket.Public,
			Key:    key,
		}
		if err := u.deps.Storage.Delete(ctx, location); err != nil {
			return err
		}
	}
	return nil
}

func (u *usecase) getUploadUrl() string {
	baseURL := strings.TrimSuffix(u.deps.Link.RouteProvider.BaseURL(), "/")
	return fmt.Sprintf("%s/%s", baseURL, u.deps.Bucket.Public)
}

func (u *usecase) ensureObjectExists(ctx context.Context, key string) error {
	loc := Location{
		Bucket: u.deps.Bucket.Public,
		Key:    key,
	}

	exists, err := u.deps.Storage.Exists(ctx, loc)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	if !exists {
		return pkg.ErrNotFound
	}

	return nil
}
