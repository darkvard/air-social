package handler

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/internal/transport/http/dto"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type AuthHandler struct {
	authSvc service.AuthService
	url     domain.URLFactory
}

func NewAuthHandler(authSvc service.AuthService, url domain.URLFactory) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		url:     url,
	}
}

// Register godoc
//
//	@Summary		Register a new user account
//	@Description	Create a new user account. Sends a verification email with a random token to the registered email address.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Register Request"
//	@Success		201		{object}	dto.UserResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		409		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.RegisterParams{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	result, err := h.authSvc.Register(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Created(c, h.mapUserToResponse(result))
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user credentials. Returns a JWT Access Token and a Refresh Token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Login Request"
//	@Success		200		{object}	dto.LoginResponse	"Returns user info and tokens"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.LoginParams{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceID,
	}

	user, token, err := h.authSvc.Login(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, dto.LoginResponse{
		User:  h.mapUserToResponse(user),
		Token: dto.NewTokenResponse(token),
	})
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Use a valid Refresh Token to obtain a new pair of JWT Access/Refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest	true	"Refresh Request"
//	@Success		200		{object}	dto.TokenResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	res, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, dto.NewTokenResponse(res))
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Revoke current device session or all sessions
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.LogoutRequest	true	"Logout Request"
//	@Success		200		{string}	string				"logout success"
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req)

	claims, err := middleware.GetAuthClaims(c)
	if err != nil || claims.UserID < 0 || claims.DeviceID == "" {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	tokenMeta, err := middleware.GetTokenMeta(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	params := domain.LogoutParams{
		UserID:       claims.UserID,
		DeviceID:     claims.DeviceID,
		IsAllDevices: req.IsAllDevices,
		Token:        tokenMeta,
	}

	if err := h.authSvc.Logout(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.SuccessWithMsg(c, "logout success", nil)
}

// VerifyEmail godoc
//
//	@Summary		Verify email address
//	@Description	Verify user email address using the random token sent during registration.
//	@Tags			Auth
//	@Produce		html
//	@Param			token	query		string	true	"Random Verification Token"
//	@Success		200		{string}	string	"HTML Page"
//	@Failure		400		{string}	string	"HTML Page"
//	@Router			/auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "verification.gohtml", gin.H{"Success": false})
		return
	}

	if err := h.authSvc.VerifyEmail(c.Request.Context(), token); err != nil {
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
//	@Param			request	body		dto.ForgotPasswordRequest	true	"Forgot Password Request"
//	@Success		200		{string}	string						"Instruction message"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	h.authSvc.ForgotPassword(c.Request.Context(), req.Email)

	pkg.SuccessWithMsg(c, "If the email exists, we have sent instructions on how to reset your password.", nil)
}

// ShowResetPasswordPage godoc
//
//	@Summary		Show reset password page
//	@Description	Render the HTML page for resetting password using the random token from email.
//	@Tags			Auth
//	@Produce		html
//	@Param			token	query		string	true	"Random Reset Token"
//	@Success		200		{string}	string	"HTML Page"
//	@Failure		400		{string}	string	"HTML Page"
//	@Router			/auth/reset-password [get]
func (h *AuthHandler) ShowResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	if !h.authSvc.IsResetPasswordTokenValid(c.Request.Context(), token) {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	c.HTML(200, "reset_password.gohtml", gin.H{
		"Success": true,
		"ApiUrl":  h.url.ResetPasswordApiURL(),
	})
}

// ResetPassword godoc
//
//	@Summary		Reset password
//	@Description	Update the user's password using the valid random token received via email.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ResetPasswordRequest	true	"Reset Password Request"
//	@Success		200		{string}	string						"password update successfully"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := domain.ResetPasswordParams{
		EmailToken: req.Token,
		Password:   req.Password,
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "password update successfully", nil)
}

func (h *AuthHandler) mapUserToResponse(user *domain.User) dto.UserResponse {
	avatar := h.url.PublicFileURL(user.Profile.Avatar)
	cover := h.url.PublicFileURL(user.Profile.CoverImage)
	return dto.NewUserResponse(user, avatar, cover)
}
