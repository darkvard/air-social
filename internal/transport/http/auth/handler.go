package auth

import (
	"errors"

	"github.com/gin-gonic/gin"

	"air-social/internal/domain/auth"
	"air-social/internal/domain/common"
	"air-social/internal/domain/user"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

type Handler struct {
	provider common.LinkProvider
	usecase  auth.UseCase
}

func NewHandler(provider common.LinkProvider, usecase auth.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}

// Register godoc
//
//	@Summary		Register a new user account
//	@Description	Create a new user account. Sends a verification email with a random token to the registered email address.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Register Request"
//	@Success		201		{object}	RegisterResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		409		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/register [post]
func (h Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := auth.RegisterParams{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	user, err := h.usecase.Register(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, pkg.ErrAlreadyExists) {
			err = pkg.NewError(err, "email or username already exists")
		}
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.Created(c, h.toRegisterResponse(*user))
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Authenticate user credentials. Returns a JWT Access Token and a Refresh Token.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Login Request"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/login [post]
func (h Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := auth.LoginParams{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceID,
	}

	user, token, err := h.usecase.Login(c.Request.Context(), params)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toLoginResponse(*user, token))
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Use a valid Refresh Token to obtain a new pair of JWT Access/Refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RefreshTokenRequest	true	"Refresh Token Request"
//	@Success		200		{object}	TokenResponse
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/refresh-token [post]
func (h Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	tokens, err := h.usecase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, h.toTokenResponse(tokens))
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Revoke current device session or all sessions
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		LogoutRequest	true	"Logout Request"
//	@Success		204		{object}	nil
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/logout [post]
func (h Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)

	claims, err := middleware.GetTokenClaims(c)
	if err != nil || claims.DeviceID == "" {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	token, err := middleware.GetAccessToken(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	params := auth.LogoutParams{
		UserID:       claims.UserID,
		DeviceID:     claims.DeviceID,
		IsAllDevices: req.IsAllDevices,
		Token:        token.Token,
		ExpiresAt:    token.ExpiresAt,
	}

	if err := h.usecase.Logout(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}
	pkg.NoContent(c)
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
func (h Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "verification.gohtml", gin.H{"Success": false})
		return
	}

	if err := h.usecase.VerifyEmail(c.Request.Context(), token); err != nil {
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
//	@Param			request	body		ForgotPasswordRequest	true	"Forgot Password Request"
//	@Success		200		{object}	pkg.Response			"Instruction message"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/forgot-password [post]
func (h Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}
	h.usecase.ForgotPassword(c.Request.Context(), req.Email)

	pkg.SuccessWithMsg(c, "If the email exists, a reset link has been sent", nil)
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
func (h Handler) ShowResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	if !h.usecase.ValidateResetPasswordToken(c.Request.Context(), token) {
		c.HTML(400, "reset_password.gohtml", gin.H{"Success": false})
		return
	}

	c.HTML(200, "reset_password.gohtml", gin.H{
		"Success": true,
		"ApiUrl":  h.provider.ResetPasswordEndpoint(),
	})
}

// ResetPassword godoc
//
//	@Summary		Reset password
//	@Description	Update the user's password using the valid random token received via email.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ResetPasswordRequest	true	"Reset Password Request"
//	@Success		200		{object}	pkg.Response			"password update successfully"
//	@Failure		400		{object}	pkg.ValidationResult
//	@Failure		404		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/auth/reset-password [post]
func (h Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	params := auth.ResetPasswordParams{
		EmailToken: req.Token,
		Password:   req.Password,
	}

	if err := h.usecase.ResetPassword(c.Request.Context(), params); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.SuccessWithMsg(c, "password has been reset successfully", nil)
}

func (h Handler) toRegisterResponse(user user.User) RegisterResponse {
	return RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Verified:  user.Status.Verified,
		CreatedAt: user.CreatedAt,
	}
}

func (h Handler) toLoginResponse(user user.User, token auth.TokenResult) LoginResponse {
	return LoginResponse{
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Verified:  user.Status.Verified,
			Avatar:    h.provider.PublicFile(user.Profile.Avatar),
			CreatedAt: user.CreatedAt,
		},
		Token: TokenResponse{
			Type:            token.Type,
			AccessToken:     token.AccessToken,
			AccessExpireAt:  token.AccessExpireAt,
			RefreshToken:    token.RefreshToken,
			RefreshExpireAt: token.RefreshExpireAt,
		},
	}
}

func (h Handler) toTokenResponse(token auth.TokenResult) TokenResponse {
	return TokenResponse{
		Type:            token.Type,
		AccessToken:     token.AccessToken,
		AccessExpireAt:  token.AccessExpireAt,
		RefreshToken:    token.RefreshToken,
		RefreshExpireAt: token.RefreshExpireAt,
	}

}
