package domain

import (
	"context"
	"mime/multipart"
)

type FileStorage interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error)

	DeleteFile(ctx context.Context, path string) error
}
