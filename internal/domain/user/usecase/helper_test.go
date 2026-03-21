package usecase_test

import (
	"testing"
	"time"

	appcache "air-social/internal/cache"
	cachemocks "air-social/internal/cache/mocks"
	"air-social/internal/domain/user"
)

// newTestCache creates a TieredCache backed by a real MemCache (L1) and a mock Cache (L2).
// The returned mockL2 can be used to set expectations on cache interactions.
func newTestCache(t *testing.T) (*cachemocks.MockCache[*user.UserSummary], appcache.TieredStore[*user.UserSummary]) {
	t.Helper()
	mockL2 := cachemocks.NewMockCache[*user.UserSummary](t)
	l1 := appcache.NewMemCache[*user.UserSummary](10, time.Minute)
	userCache := appcache.NewTieredCache(l1, mockL2, time.Minute, 12*time.Hour)
	return mockL2, userCache
}
