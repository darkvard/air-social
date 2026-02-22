package media

import (
	"time"

	"air-social/internal/domain/media"
)

// PresignRequest represents the metadata for a single file upload request.
type PresignRequest struct {
	FileName string              `json:"file_name" binding:"required"`
	FileType string              `json:"file_type" binding:"required"`
	FileSize int64               `json:"file_size" binding:"required"`
	Domain   media.UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature  media.UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

// BulkPresignRequest represents a request to generate multiple presigned URLs.
type BulkPresignRequest struct {
	Files []PresignRequest `json:"files" binding:"required,min=1,max=10,dive"`
}

// ConfirmRequest represents the request to confirm a successful file upload.
type ConfirmRequest struct {
	ObjectKey string              `json:"object_key" binding:"required"`
	Domain    media.UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature   media.UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

// PresignResponse contains the presigned URL and form data required for the client to upload the file.
type PresignResponse struct {
	FileName  string            `json:"file_name"`
	UploadURL string            `json:"upload_url"`
	FormData  map[string]string `json:"form_data"`
	ObjectKey string            `json:"object_key"`
	PublicURL string            `json:"public_url"`
	ExpireAt  time.Time         `json:"expire_at"`
}

// ConfirmResponse contains the public URL of the confirmed file.
type ConfirmResponse struct {
	Domain  media.UploadDomain  `json:"domain"`
	Feature media.UploadFeature `json:"feature"`
	URL     string              `json:"url"`
}
