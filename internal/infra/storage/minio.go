package storage

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
)

type minioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) *minioStorage {
	return &minioStorage{
		client: client,
	}
}

func (m *minioStorage) GetPresignedPutURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, bucket, objectName, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (m *minioStorage) StatFile(ctx context.Context, bucket, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *minioStorage) DeleteFile(ctx context.Context, bucket, objectName string) error {
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}
