package minio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"air-social/internal/config"
	"air-social/pkg"
)

func NewConnection(cfg config.MinioStorageConfig) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UserSSl,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: invalid config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = pkg.Retry(ctx, 10, 2*time.Second, func() error {
		_, err := minioClient.ListBuckets(ctx)
		return err
	}); err != nil {
		return nil, fmt.Errorf("minio: connection failed after retries: %w", err)
	}

	buckets := []string{cfg.BucketPublic, cfg.BucketPrivate}
	for _, bucketName := range buckets {
		err := ensureBucketExists(ctx, minioClient, bucketName, cfg.BucketPublic)
		if err != nil {
			return nil, err
		}
	}

	return minioClient, nil
}

func ensureBucketExists(ctx context.Context, client *minio.Client, bucketName, publicName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket %s error: %w", bucketName, err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %s error: %w", bucketName, err)
		}

		if bucketName == publicName {
			policy := fmt.Sprintf(`{
                "Version": "2012-10-17",
                "Statement": [
                    {"Effect": "Allow", "Principal": {"AWS": ["*"]}, "Action": ["s3:GetObject"], "Resource": ["arn:aws:s3:::%s/*"]}
                ]
            }`, bucketName)

			if err := client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
				return fmt.Errorf("set public policy error: %w", err)
			}
			fmt.Printf("Set public policy for: %s\n", bucketName)
		}
	}

	return nil
}

func NewMinioStorage(client *minio.Client) (*minioStorage, error) {
	if client == nil {
		return nil, errors.New("minio client cannot nil")
	}
	return newMinioStorage(client), nil
}