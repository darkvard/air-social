package media

import "time"

type PresignParams struct {
	EntityID int64
	FileName string
	FileType string
	FileSize int64
	Domain   UploadDomain
	Feature  UploadFeature
}

type ConfirmParams struct {
	EntityID  int64
	ObjectKey string
	Domain    UploadDomain
	Feature   UploadFeature
}

type PresignedResult struct {
	FileName  string
	UploadURL string
	FormData  map[string]string
	ObjectKey string
	PublicURL string
	ExpireAt  time.Time
}

type ConfirmResult struct {
	Domain  UploadDomain
	Feature UploadFeature
	URL     string
}
