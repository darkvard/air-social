package dto

import (
	"time"

	"air-social/internal/domain"
)

type PresignedUploadRequest struct {
	FileName string               `json:"file_name" binding:"required"`
	FileType string               `json:"file_type" binding:"required"`
	FileSize int64                `json:"file_size" binding:"required"`
	Domain   domain.UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature  domain.UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

type BulkPresignedUploadRequest struct {
	Files []PresignedUploadRequest `json:"files" binding:"required,min=1,max=10,dive"`
}

type ConfirmUploadRequest struct {
	ObjectKey string               `json:"object_key" binding:"required"`
	Domain    domain.UploadDomain  `json:"domain" binding:"required,oneof=users posts groups messages"`
	Feature   domain.UploadFeature `json:"feature" binding:"required,oneof=avatar cover feed_image feed_video voice_chat attachment"`
}

type PresignedFileResponse struct {
	FileName  string            `json:"file_name"`
	UploadURL string            `json:"upload_url"`
	FormData  map[string]string `json:"form_data"`
	ObjectKey string            `json:"object_key"`
	PublicURL string            `json:"public_url"`
	ExpireAt  time.Time         `json:"expire_at"`
}

type ConfirmFileResponse struct {
	Domain  domain.UploadDomain  `json:"domain"`
	Feature domain.UploadFeature `json:"feature"`
	URL     string               `json:"url"`
}
