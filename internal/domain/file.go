package domain

import (
	"context"
	"time"
)

const (
	DomainUser    UploadDomain = "users"
	DomainPost    UploadDomain = "posts"
	DomainGroup   UploadDomain = "groups"
	DomainMessage UploadDomain = "messages"
)

const (
	FeatureAvatar     UploadFeature = "avatar"
	FeatureCover      UploadFeature = "cover"
	FeatureFeedImage  UploadFeature = "feed_image"
	FeatureFeedVideo  UploadFeature = "feed_video"
	FeatureVoiceChat  UploadFeature = "voice_chat"
	FeatureAttachment UploadFeature = "attachment"
)

const (
	Limit5MB   int64 = 5 * 1024 * 1024
	Limit10MB  int64 = 10 * 1024 * 1024
	Limit50MB  int64 = 50 * 1024 * 1024
	Limit100MB int64 = 100 * 1024 * 1024
)

const (
	UploadExpiry = 15 * time.Minute
)

var (
	ImageAllowedTypes = []string{"image/jpeg", "image/png", "image/webp", "image/jpg", "image/gif"}
	VideoAllowedTypes = []string{"video/mp4", "video/quicktime", "video/webm"}
	AudioAllowedTypes = []string{"audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4"}
)

type FileStorage interface {
	// Assuming HTTP for internal communication within Docker network. e.g. "minio:9000"
	GetEndpoint() string
	GetPresignedPostPolicy(ctx context.Context, loc StorageLocation, constraints UploadConstraints) (PresignedURLResult, error)
	StatFile(ctx context.Context, loc StorageLocation) (bool, error)
	DeleteFile(ctx context.Context, loc StorageLocation) error
}

type UploadDomain string

type UploadFeature string

type UploadRule struct {
	MaxBytes     int64
	AllowedTypes []string // Mime Types
}

type FileConfig struct {
	PublicPathPrefix string // e.g. "http://localhost/air-social-public" (via Nginx)
	BucketPublic     string
	BucketPrivate    string
}

type StorageLocation struct {
	Bucket string
	Key    string
}

type UploadConstraints struct {
	Expiry      time.Duration
	ContentType string
	MaxSize     int64
}

type PresignedURLResult struct {
	UploadURL string
	FormData  map[string]string
}

type PresignedFileParams struct {
	UserID   int64
	FileName string
	FileType string
	FileSize int64
	Domain   UploadDomain
	Feature  UploadFeature
}

type ConfirmFileParams struct {
	UserID    int64
	ObjectKey string
	Domain    UploadDomain
	Feature   UploadFeature
}

type PresignedFileUploadRequest struct {
	FileName string        `json:"file_name" binding:"required"`
	FileType string        `json:"file_type" binding:"required"`
	FileSize int64         `json:"file_size" binding:"required"`
	Domain   UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature  UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

type ConfirmFileUploadRequest struct {
	ObjectKey string        `json:"object_key" binding:"required"`
	Domain    UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature   UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

type PresignedFileResponse struct {
	UploadURL     string            `json:"upload_url"`
	FormData      map[string]string `json:"form_data"`
	ObjectKey     string            `json:"object_key"`
	PublicURL     string            `json:"public_url"`
	ExpirySeconds int64             `json:"expiry_seconds"`
}
