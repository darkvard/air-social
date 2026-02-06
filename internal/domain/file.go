package domain

import (
	"context"
	"time"
)

type UploadDomain string
type UploadFeature string

const (
	DomainUser    UploadDomain = "users"
	DomainPost    UploadDomain = "posts"
	DomainMessage UploadDomain = "messages"

	FeatureAvatar     UploadFeature = "avatar"
	FeatureCover      UploadFeature = "cover"
	FeatureFeedImage  UploadFeature = "feed_image"
	FeatureFeedVideo  UploadFeature = "feed_video"
	FeatureVoiceChat  UploadFeature = "voice_chat"
	FeatureAttachment UploadFeature = "attachment"
)
const (
	PresignedUploadExpiry = 30 * time.Minute

	Limit5MB   int64 = 5 * 1024 * 1024
	Limit10MB  int64 = 10 * 1024 * 1024
	Limit50MB  int64 = 50 * 1024 * 1024
	Limit100MB int64 = 100 * 1024 * 1024
)

var (
	ImageAllowedTypes = []string{"image/jpeg", "image/png", "image/webp", "image/jpg", "image/gif"}
	VideoAllowedTypes = []string{"video/mp4", "video/quicktime", "video/webm"}
	AudioAllowedTypes = []string{"audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4"}
)

var FileUploadRules = map[UploadDomain]map[UploadFeature]UploadRule{
	DomainUser: {
		FeatureAvatar: {MaxBytes: Limit5MB, AllowedTypes: ImageAllowedTypes},
		FeatureCover:  {MaxBytes: Limit5MB, AllowedTypes: ImageAllowedTypes},
	},
	DomainPost: {
		FeatureFeedImage: {MaxBytes: Limit10MB, AllowedTypes: ImageAllowedTypes},
		FeatureFeedVideo: {MaxBytes: Limit100MB, AllowedTypes: VideoAllowedTypes},
	},
	DomainMessage: {
		FeatureVoiceChat:  {MaxBytes: Limit10MB, AllowedTypes: AudioAllowedTypes},
		FeatureFeedImage:  {MaxBytes: Limit10MB, AllowedTypes: ImageAllowedTypes},
		FeatureAttachment: {MaxBytes: Limit50MB, AllowedTypes: nil},
	},
}

var ValidDomainFeatures = map[UploadDomain][]UploadFeature{
	DomainUser:    {FeatureAvatar, FeatureCover},
	DomainPost:    {FeatureFeedImage, FeatureFeedVideo},
	DomainMessage: {FeatureAttachment, FeatureVoiceChat, FeatureFeedImage},
}

type FileStorage interface {
	GetEndpoint() string
	GetPresignedPostPolicy(ctx context.Context, loc StorageLocation, constraints UploadConstraints) (PresignedURLResult, error)
	StatFile(ctx context.Context, loc StorageLocation) (bool, error)
	DeleteFile(ctx context.Context, loc StorageLocation) error
}

type UploadRule struct {
	MaxBytes     int64
	AllowedTypes []string // Mime Types
}

type FileConfig struct {
	BucketPublic  string
	BucketPrivate string
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
	EntityID int64
	FileName string
	FileType string
	FileSize int64
	Domain   UploadDomain
	Feature  UploadFeature
}

type ConfirmFileParams struct {
	EntityID  int64
	ObjectKey string
	Domain    UploadDomain
	Feature   UploadFeature
}

type PresignedFile struct {
	UploadURL string
	FormData  map[string]string
	ObjectKey string
	PublicURL string
	ExpireAt  time.Time
}
