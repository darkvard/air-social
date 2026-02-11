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

func (r *BulkPresignedUploadRequest) ToDomain(userID int64) []domain.PresignedFileParams {
	params := make([]domain.PresignedFileParams, len(r.Files))
	for i, f := range r.Files {
		params[i] = domain.PresignedFileParams{
			EntityID: userID,
			FileName: f.FileName,
			FileType: f.FileType,
			FileSize: f.FileSize,
			Domain:   f.Domain,
			Feature:  f.Feature,
		}
	}
	return params
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

func NewPresignedFileResponseList(files []domain.PresignedFile) []PresignedFileResponse {
	resp := make([]PresignedFileResponse, len(files))
	for i, r := range files {
		resp[i] = PresignedFileResponse{
			FileName:  r.FileName,
			UploadURL: r.UploadURL,
			FormData:  r.FormData,
			ObjectKey: r.ObjectKey,
			PublicURL: r.PublicURL,
			ExpireAt:  r.ExpireAt,
		}
	}
	return resp
}

type ConfirmFileResponse struct {
	Domain  domain.UploadDomain  `json:"domain"`
	Feature domain.UploadFeature `json:"feature"`
	URL     string               `json:"url"`
}

func NewConfirmFileResponseList(results []domain.ConfirmFileResult) []ConfirmFileResponse {
	resp := make([]ConfirmFileResponse, len(results))
	for i, res := range results {
		resp[i] = ConfirmFileResponse{
			Domain:  res.Domain,
			Feature: res.Feature,
			URL:     res.URL,
		}
	}
	return resp
}
