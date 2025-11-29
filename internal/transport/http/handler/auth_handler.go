package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/pkg"
)

type AuthHandler struct {
	auth service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.auth.Register(c.Request.Context(), &req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, result)
}
