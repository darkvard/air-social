package health

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/health"
	"air-social/pkg"
)

type Handler struct {
	usecase health.UseCase
}

func NewHandler(usecase health.UseCase) Handler {
	return Handler{
		usecase: usecase,
	}
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Check the health status of the application components
//	@Tags			Health
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	pkg.Response
//	@Router			/health [get]
func (h Handler) HealthCheck(c *gin.Context) {
	_, details := h.usecase.CheckStatus(c.Request.Context())
	pkg.Success(c, details)
}

// Welcome godoc
//
//	@Summary		Welcome page
//	@Description	Render the API welcome page with status overview.
//	@Tags			Health
//	@Produce		html
//	@Success		200	{string}	string	"HTML Page"
//	@Failure		503	{string}	string	"HTML Page (Maintenance)"
//	@Router			/ [get]
func (h Handler) Welcome(c *gin.Context) {
	resp := h.usecase.Overview(c.Request.Context())
	c.HTML(resp.HTTPCode, "welcome.gohtml", resp)
}
