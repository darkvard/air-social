package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"

	"air-social/internal/domain"
)

type minioStorage struct {
	client *minio.Client
}

func newMinioStorage(client *minio.Client) *minioStorage {
	return &minioStorage{
		client: client,
	}
}

func (m *minioStorage) GetPresignedPostPolicy(ctx context.Context, loc domain.StorageLocation, constraints domain.UploadConstraints) (domain.PresignedURLResult, error) {
	empty := domain.PresignedURLResult{}

	policy, err := m.setupPostPolicy(loc, constraints)
	if err != nil {
		return empty, err
	}

	url, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return empty, err
	}

	return domain.PresignedURLResult{
		UploadURL: url.String(),
		FormData:  formData,
	}, nil
}

func (m *minioStorage) setupPostPolicy(loc domain.StorageLocation, constraints domain.UploadConstraints) (*minio.PostPolicy, error) {
	policy := minio.NewPostPolicy()
	if err := policy.SetBucket(loc.Bucket); err != nil {
		return nil, err
	}
	if err := policy.SetKey(loc.Key); err != nil {
		return nil, err
	}
	if err := policy.SetExpires(time.Now().UTC().Add(constraints.Expiry)); err != nil {
		return nil, err
	}
	if err := policy.SetContentType(constraints.ContentType); err != nil {
		return nil, err
	}
	if err := policy.SetContentLengthRange(1024, constraints.MaxSize); err != nil {
		return nil, err
	}
	return policy, nil
}

func (m *minioStorage) StatFile(ctx context.Context, loc domain.StorageLocation) (bool, error) {
	_, err := m.client.StatObject(ctx, loc.Bucket, loc.Key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *minioStorage) DeleteFile(ctx context.Context, loc domain.StorageLocation) error {
	return m.client.RemoveObject(ctx, loc.Bucket, loc.Key, minio.RemoveObjectOptions{})
}

func (s *minioStorage) GetEndpoint() string {
	return fmt.Sprintf("http://%s", s.client.EndpointURL().Host)
}
