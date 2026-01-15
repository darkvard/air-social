package domain

import (
	"context"
	"time"
)

const (
	AvatarType FileType = "avatar"
	CoverType  FileType = "cover"
)

const (
	UploadExpiry = 30 * time.Minute
)

type FileStorage interface {
	GetPresignedPutURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)

	StatFile(ctx context.Context, bucket, objectName string) (bool, error)

	DeleteFile(ctx context.Context, bucket, objectName string) error
}

type PresignedFile struct {
	UserID int64
	Ext    string   // .jpg, .png ...
	Folder string   // users, posts, chats ...
	Typ    FileType // avatar, cover ...
}

type ConfirmFile struct {
	UserID     int64
	ObjectName string
	Typ        FileType
}

type FileConfig struct {
	Env           string
	PublicURL     string
	BucketPublic  string
	BucketPrivate string
}

type FileType string

type PresignedFileUploadRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileType string `json:"file_type" binding:"required,oneof=avatar cover"` // Validate enum
}

type ConfirmFileUploadRequest struct {
	ObjectName string `json:"object_name" binding:"required"`
	FileType   string `json:"file_type" binding:"required,oneof=avatar cover"`
}

type PresignedFileResponse struct {
	UploadURL     string `json:"upload_url"`
	ObjectName    string `json:"object_name"`
	ExpirySeconds int64  `json:"expiry_seconds"`
}
