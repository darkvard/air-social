package domain

import (
	"context"
	"fmt"
	"time"
)

// <system>:<feature>:<state>:<id>
const (
	WorkerEmailProcessedKey = "worker:email:processed:"
	WorkerEmailVerifyKey    = "worker:email:verify:"
	WorkerEmailResetKey     = "worker:email:reset:"
	WorkerEmailRetryKey     = "worker:email:retry:"

	UserUploadImageVerifyKey = "user:upload:verify:"
	UserBlacklistTokenKey    = "user:blacklist:token:"
	UserFollowerCountKey     = "user:followers:count:"
	UserFollowingCountKey    = "user:followings:count:"
	UserSummaryKey           = "user:public_info:"
)

type CacheStorage interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	IsExist(ctx context.Context, key string) (bool, error)
}

func GetEmailVerificationKey(token string) string {
	return fmt.Sprintf(WorkerEmailVerifyKey+"%s", token)
}

func GetEmailProcessedKey(token string) string {
	return fmt.Sprintf(WorkerEmailProcessedKey+"%s", token)
}

func GetEmailResetPasswordKey(token string) string {
	return fmt.Sprintf(WorkerEmailResetKey+"%s", token)
}

func GetEmailRetryKey(token string) string {
	return fmt.Sprintf(WorkerEmailRetryKey+"%s", token)
}

func GetUploadImageKey(objectName string) string {
	return fmt.Sprintf(UserUploadImageVerifyKey+"%s", objectName)
}

func GetBlacklistTokenKey(token string) string {
	return fmt.Sprintf(UserBlacklistTokenKey+"%s", token)
}

func GetFollowerCountKey(userID int64) string {
	return fmt.Sprintf(UserFollowerCountKey+"%d", userID)
}

func GetFollowingCountKey(userID int64) string {
	return fmt.Sprintf(UserFollowingCountKey+"%d", userID)
}

func GetUserSummaryKey(userID int64) string {
	return fmt.Sprintf(UserSummaryKey+"%d", userID)
}
