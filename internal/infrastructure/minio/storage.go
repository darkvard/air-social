package minio

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"

	"air-social/internal/domain/media"
	"air-social/pkg"
)

type storage struct {
	client *minio.Client
}

func NewStorage(client *minio.Client) (*storage, error) {
	if client == nil {
		return nil, errors.New("minio client cannot nil")
	}
	return &storage{
		client: client,
	}, nil
}

func (m *storage) PresignUpload(ctx context.Context, location media.Location, constraints media.Constraints) (media.UploadForm, error) {
	empty := media.UploadForm{}

	policy, err := m.setupPostPolicy(location, constraints)
	if err != nil {
		return empty, err
	}

	url, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return empty, err
	}

	return media.UploadForm{
		URL:    url.String(),
		Fields: formData,
	}, nil
}

func (m *storage) setupPostPolicy(location media.Location, constraints media.Constraints) (*minio.PostPolicy, error) {
	policy := minio.NewPostPolicy()
	if err := policy.SetBucket(location.Bucket); err != nil {
		return nil, err
	}
	if err := policy.SetKey(location.Key); err != nil {
		return nil, err
	}
	if err := policy.SetExpires(pkg.TimeNowUTC().Add(constraints.Expiry)); err != nil {
		return nil, err
	}
	if err := policy.SetContentType(constraints.ContentType); err != nil {
		return nil, err
	}
	if err := policy.SetContentLengthRange(1, constraints.MaxSize); err != nil {
		return nil, err
	}
	return policy, nil
}

func (m *storage) Exists(ctx context.Context, location media.Location) (bool, error) {
	_, err := m.client.StatObject(ctx, location.Bucket, location.Key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *storage) Delete(ctx context.Context, location media.Location) error {
	return m.client.RemoveObject(ctx, location.Bucket, location.Key, minio.RemoveObjectOptions{})
}
