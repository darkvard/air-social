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

	UserUploadImageVerify = "user:upload:verify:"
	UserBlacklistToken    = "user:blacklist:token:"
	UserFollowerCount     = "user:followers:count:"
	UserFollowingCount    = "user:followings:count:"
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
	return fmt.Sprintf(UserUploadImageVerify+"%s", objectName)
}

func GetBlacklistTokenKey(token string) string {
	return fmt.Sprintf(UserBlacklistToken+"%s", token)
}

func GetFollowerCountKey(userID int64) string {
	return fmt.Sprintf(UserFollowerCount+"%d", userID)
}

func GetFollowingCountKey(userID int64) string {
	return fmt.Sprintf(UserFollowingCount+"%d", userID)
}
