package url

import (
	"fmt"

	"air-social/internal/domain/common"
)

type system struct {
	route common.RouteProvider
}

func newSystemProvider(route common.RouteProvider) *system {
	return &system{route: route}
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
