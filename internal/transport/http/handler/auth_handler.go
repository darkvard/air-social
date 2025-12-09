package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/middleware"
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	user, token, err := h.auth.Login(c.Request.Context(), &req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), &req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	payload, err := middleware.GetAuthPayload(c)
	if err != nil || payload.UserID < 0 || payload.DeviceID == "" {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	req.UserID = payload.UserID
	req.DeviceID = payload.DeviceID

	if err := h.auth.Logout(c.Request.Context(), &req); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.Success(c, "logout success")
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {

}

func (h *AuthHandler) ResetPassword(c *gin.Context) {

}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {

}
