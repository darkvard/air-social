package handler

import (
	"errors"

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

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "verification.gohtml", gin.H{"Success": false})
		return
	}

	if err := h.auth.VerifyEmail(c.Request.Context(), token); err != nil {
		c.HTML(400, "verification.gohtml", gin.H{"Success": false})
		return
	}

	c.HTML(200, "verification.gohtml", gin.H{"Success": true})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if err := h.auth.ForgotPassword(c.Request.Context(), &req); err != nil {
		if errors.Is(err, pkg.ErrInternal) {
			pkg.Log().Errorw("failed to forgot password", "error", err)
		}
	}

	pkg.Success(c, "If the email exists, we have sent instructions on how to reset your password.")
}

func (h *AuthHandler) ShowResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	if err := h.auth.ResetPassword(
		c.Request.Context(),
		&domain.ResetPasswordRequest{Token: token, Password: ""},
		true,
	); err != nil {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	c.HTML(200, "reset_password.gohtml", gin.H{"Success": true})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if err := h.auth.ResetPassword(c.Request.Context(), &req, false); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, "password update successfully")
}
