package media

import "time"

// PresignParams contains the parameters required to generate a presigned URL.
type PresignParams struct {
	EntityID int64
	FileName string
	FileType string
	FileSize int64
	Domain   UploadDomain
	Feature  UploadFeature
}

// ConfirmParams contains the parameters required to confirm a file upload.
type ConfirmParams struct {
	EntityID  int64
	ObjectKey string
	Domain    UploadDomain
	Feature   UploadFeature
}

// PresignedResult represents the result of a presigned URL generation.
type PresignedResult struct {
	FileName  string
	UploadURL string
	FormData  map[string]string
	ObjectKey string
	PublicURL string
	ExpireAt  time.Time
}

// ConfirmResult represents the result of a file confirmation.
type ConfirmResult struct {
	Domain  UploadDomain
	Feature UploadFeature
	URL     string
}
