package url

import (
	"fmt"
	"strings"

	"air-social/internal/config"
	"air-social/internal/domain/common"
)

type link struct {
	route        common.RouteProvider
	bucketPublic string
}

func newLinkProvider(route common.RouteProvider, cfg config.MinioStorageConfig) *link {
	return &link{
		route:        route,
		bucketPublic: cfg.BucketPublic,
	}
}

func (u *link) VerifyEmail(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", u.getApiPath(), "/auth", "/verify-email", token)
}

func (u *link) ResetPassword(token string) string {
	return fmt.Sprintf("%s%s%s?token=%s", u.getApiPath(), "/auth", "/reset-password", token)
}

func (u *link) ResetPasswordEndpoint() string {
	return fmt.Sprintf("%s%s%s", u.getApiPath(), "/auth", "/reset-password")
}

func (u *link) PublicFile(key string) string {
	if key == "" {
		return ""
	}
	domain := strings.TrimSuffix(u.route.BaseURL(), "/")
	return fmt.Sprintf("%s/%s/%s", domain, u.bucketPublic, key)
}

func (u *link) getApiPath() string {
	return u.route.ApiPath()
}
