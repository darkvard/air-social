package media

import "context"

type Storage interface {
	Exists(ctx context.Context, location Location) (bool, error)
	Delete(ctx context.Context, location Location) error
	PresignUpload(ctx context.Context, location Location, constraints Constraints) (UploadForm, error)
}

// todo: update minio