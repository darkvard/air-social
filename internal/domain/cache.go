package domain


import (
	"context"
	"fmt"
	"time"
)

// <system>:<feature>:<state>:<id>
const (
	WorkerEmailProcessed = "worker:email:processed:"
	WorkerEmailVerify    = "worker:email:verify:"
	WorkerEmailReset     = "worker:email:reset:"
	WorkerEmailRetry     = "worker:email:retry:"
	UploadImageVerify    = "upload:verify:"
)

const (
	OneMinuteTime      = 1 * time.Minute
	FiveMinutesTime    = 5 * time.Minute
	TenMinutesTime     = 10 * time.Minute
	FifteenMinutesTime = 15 * time.Minute
	ThirtyMinutesTime  = 30 * time.Minute
	OneHourTime        = 1 * time.Hour
	OneDayTime         = 24 * time.Hour
)

type CacheStorage interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	IsExist(ctx context.Context, key string) (bool, error)
}

func GetEmailVerificationKey(token string) string {
	return fmt.Sprintf(WorkerEmailVerify+"%s", token)
}

func GetEmailProcessedKey(token string) string {
	return fmt.Sprintf(WorkerEmailProcessed+"%s", token)
}

func GetEmailResetPasswordKey(token string) string {
	return fmt.Sprintf(WorkerEmailReset+"%s", token)
}

func GetEmailRetryKey(token string) string {
	return fmt.Sprintf(WorkerEmailRetry+"%s", token)
}

func GetUploadImageKey(objectName string) string {
	return fmt.Sprintf(UploadImageVerify+"%s", objectName)
}
