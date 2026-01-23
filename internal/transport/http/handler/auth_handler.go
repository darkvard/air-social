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

// Register godoc
//
//	@Summary		Register a new user account
//	@Description	Create a new user account. Sends a verification email with a random token to the registered email address.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.RegisterRequest	true	"Register Request"
//	@Success		200		{object}	domain.UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		409		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.auth.Register(c.Request.Context(), req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, result)
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user credentials. Returns a JWT Access Token and a Refresh Token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.LoginRequest		true	"Login Request"
//	@Success		200		{object}	map[string]interface{}	"Returns user info and tokens"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	user, token, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Use a valid Refresh Token to obtain a new pair of JWT Access/Refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.RefreshRequest	true	"Refresh Request"
//	@Success		200		{object}	domain.TokenInfo
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), req)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, tokens)
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Revoke current device session or all sessions
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.LogoutRequest	true	"Logout Request"
//	@Success		200		{string}	string					"logout success"
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/logout [post]
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

	if err := h.auth.Logout(c.Request.Context(), req); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.Success(c, "logout success")
}

// VerifyEmail godoc
//
//	@Summary		Verify email address
//	@Description	Verify user email address using the random token sent during registration.
//	@Tags			Auth
//	@Produce		html
//	@Param			token	query		string	true	"Random Verification Token"
//	@Success		200		{string}	string	"HTML Page"
//	@Router			/auth/verify-email [get]
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

// ForgotPassword godoc
//
//	@Summary		Request password reset
//	@Description	Initiate password reset process. Sends an email containing a random token to reset the password.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.ForgotPasswordRequest	true	"Forgot Password Request"
//	@Success		200		{string}	string							"Instruction message"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if err := h.auth.ForgotPassword(c.Request.Context(), req); err != nil {
		if !errors.Is(err, pkg.ErrNotFound) {
			pkg.Log().Errorw("failed to forgot password", "error", err)
		}
	}

	pkg.Success(c, "If the email exists, we have sent instructions on how to reset your password.")
}

// ShowResetPasswordPage godoc
//
//	@Summary		Show reset password page
//	@Description	Render the HTML page for resetting password using the random token from email.
//	@Tags			Auth
//	@Produce		html
//	@Param			token	query		string	true	"Random Reset Token"
//	@Success		200		{string}	string	"HTML Page"
//	@Router			/auth/reset-password [get]
func (h *AuthHandler) ShowResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	if err := h.auth.ResetPassword(
		c.Request.Context(),
		domain.ResetPasswordRequest{Token: token, Password: ""},
		true,
	); err != nil {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	c.HTML(200, "reset_password.gohtml", gin.H{"Success": true})
}

// ResetPassword godoc
//
//	@Summary		Reset password
//	@Description	Update the user's password using the valid random token received via email.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.ResetPasswordRequest	true	"Reset Password Request"
//	@Success		200		{string}	string						"password update successfully"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	if err := h.auth.ResetPassword(c.Request.Context(), req, false); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, "password update successfully")
}
