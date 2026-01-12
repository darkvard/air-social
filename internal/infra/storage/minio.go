package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"

	"air-social/internal/config"
)

type minioStorage struct {
	client *minio.Client
	cfg    config.FileStorageConfig
}

func NewMinioStorage(client *minio.Client, cfg config.FileStorageConfig) *minioStorage {
	return &minioStorage{
		client: client,
		cfg:    cfg,
	}
}

func (m *minioStorage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	ext := filepath.Ext(header.Filename)
	newFileName := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := m.client.PutObject(ctx, m.cfg.Bucket, newFileName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return "", err
	}

	fullURL := fmt.Sprintf("%s/%s/%s", m.cfg.PublicURL, m.cfg.Bucket, newFileName)
	return fullURL, nil
}

func (m *minioStorage) DeleteFile(ctx context.Context, path string) error {
	// e.g. path="avatar/123.jpg" (key file not full url)
	if path == "" {
		return nil
	}

	return m.client.RemoveObject(ctx, m.cfg.Bucket, path, minio.RemoveObjectOptions{})
}
