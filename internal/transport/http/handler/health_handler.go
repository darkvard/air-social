package handler

import (
	"maps"

	"github.com/gin-gonic/gin"

	"air-social/internal/service"
	"air-social/pkg"
)

type HealthHandler struct {
	srv service.HealthService
}

func NewHealthHandler(srv service.HealthService) *HealthHandler {
	return &HealthHandler{
		srv: srv,
	}
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Check the health status of the application components
//	@Tags			Health
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	_, details := h.srv.Check(c.Request.Context())
	pkg.Success(c, details)
}

func (h *HealthHandler) Welcome(c *gin.Context) {
	isHealthy, _ := h.srv.Check(c.Request.Context())
	appInfo := h.srv.GetAppInfo()

	statusStr := "Active"
	httpCode := 200

	if !isHealthy {
		statusStr = "Maintenance"
		httpCode = 503
	}

	data := gin.H{"Status": statusStr}
	maps.Copy(data, appInfo)

	c.HTML(httpCode, "welcome.gohtml", data)
}
