package url

import (
	"fmt"

	"air-social/internal/domain/common"
	"air-social/pkg"
)

type system struct {
	route common.RouteProvider
}

func newSystemProvider(route common.RouteProvider) *system {
	u := &system{route: route}
	u. Print()
	return u
}

func (u *system) SwaggerURL() string {
	return fmt.Sprintf("%s/swagger/index.html", u.route.ApiPath())
}

func (u *system) MinioConsoleURL() string {
	return fmt.Sprintf("%s/storage-admin/", u.route.BaseURL())
}

func (u *system) RabbitMQDashboardURL() string {
	return fmt.Sprintf("%s/rabbitmq/", u.route.BaseURL())
}

func (u *system) Print() {
	pkg.Log().Infow("SwaggerURL", "url", u.SwaggerURL())
	pkg.Log().Infow("MinioConsoleURL", "url", u.MinioConsoleURL())
	pkg.Log().Infow("RabbitMQDashboardURL", "url", u.RabbitMQDashboardURL())
}
