package media

import (
	"time"
)

const (
	DomainUser    UploadDomain = "users"
	DomainPost    UploadDomain = "posts"
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

const PresignedUploadExpiry = 30 * time.Minute

var (
	ImageAllowedTypes = []string{"image/jpeg", "image/png", "image/webp", "image/jpg", "image/gif"}
	VideoAllowedTypes = []string{"video/mp4", "video/quicktime", "video/webm"}
	AudioAllowedTypes = []string{"audio/mpeg", "audio/wav", "audio/ogg", "audio/mp4"}
)

var MediaUploadRules = map[UploadDomain]map[UploadFeature]Rule{
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

type UploadDomain string

type UploadFeature string
