package media

import (
	"time"
)

type Rule struct {
	MaxBytes     int64
	AllowedTypes []string // Mime Types
}

type Bucket struct {
	Public  string
	Private string
}

type Location struct {
	Bucket string
	Key    string
}

type Constraints struct {
	Expiry      time.Duration
	ContentType string
	MaxSize     int64
}

type UploadForm struct {
	URL    string
	Fields map[string]string
}
