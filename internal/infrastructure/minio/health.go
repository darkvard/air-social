package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type Health struct {
	*minio.Client
}

func NewHealth(client *minio.Client) *Health {
	return &Health{
		Client: client,
	}
}

func (h *Health) Ping(ctx context.Context) error {
	_, err := h.Client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("minio not reachable: %w", err)
	}
	return nil
}
