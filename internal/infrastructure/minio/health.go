package minio

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
)

type Health struct {
	*minio.Client
	bucketName string
}

func NewHealth(client *minio.Client, bucketName string) *Health {
	return &Health{
		Client:     client,
		bucketName: bucketName,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err := h.Client.BucketExists(ctx, h.bucketName)
	return err
}
