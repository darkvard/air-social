package auth

import (
	"context"
	"errors"

	"air-social/internal/domain/auth/token"
	"air-social/internal/domain/auth/verify"
	"air-social/internal/domain/user"
	"air-social/pkg"
)

type UseCase interface {
	Register(ctx context.Context, params RegisterParams) (*user.User, error)
	Logout(ctx context.Context, params LogoutParams) error
	Login(ctx context.Context, params LoginParams) (*user.User, TokenResult, error)

	VerifyEmail(ctx context.Context, emailToken string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, params ResetPasswordParams) error
	ValidateResetPasswordToken(ctx context.Context, token string) bool

	RefreshToken(ctx context.Context, refreshToken string) (TokenResult, error)
}

type Deps struct {
	TokenRepo      token.Repository
	TokenProvider  token.Provider
	VerifyProvider verify.Provider
	UserFetch      user.FetchUseCase
	UserAccount    user.AccountUseCase
}

type usecase struct {
	deps Deps
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{deps: deps}
}

func (u *usecase) Register(ctx context.Context, params RegisterParams) (*user.User, error) {
	arg := user.CreateParams{
		Email:       params.Email,
		Username:    params.Username,
		NewPassword: params.Password,
	}

	result, err := u.deps.UserAccount.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	_ = u.deps.VerifyProvider.SendVerification(ctx, arg.Email, arg.Username)

	return result, nil
}

func (u *usecase) Logout(ctx context.Context, params LogoutParams) error {
	if u.deps.TokenProvider.IsBlacklisted(ctx, params.Token) {
		return pkg.ErrUnauthorized
	}

	var err error
	if params.IsAllDevices {
		err = u.deps.TokenRepo.UpdateRevokedByUser(ctx, params.UserID)
	} else {
		err = u.deps.TokenRepo.UpdateRevokedByDevice(ctx, params.UserID, params.DeviceID)
	}
	if err != nil {
		return pkg.OrInternalError(err)
	}

	u.deps.TokenProvider.AddToBlacklist(ctx, params.Token, params.ExpiresAt)
	return nil
}

func (u *usecase) Login(ctx context.Context, params LoginParams) (*user.User, TokenResult, error) {
	var emptyUser *user.User
	var emptyToken TokenResult

	account, err := u.deps.UserAccount.Authenticate(ctx, user.AuthenticateParams{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err, pkg.ErrInvalidCredentials)
	}

	refreshToken := u.deps.TokenProvider.GenerateRefreshToken()
	accessToken, err := u.deps.TokenProvider.GenerateAccessToken(account.ID, params.DeviceID)
	if err != nil {
		return emptyUser, emptyToken, pkg.ErrInternal
	}

	if err = u.deps.TokenRepo.Create(ctx, &token.RefreshToken{
		UserID:    account.ID,
		DeviceID:  params.DeviceID,
		TokenHash: refreshToken.Hashed,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: pkg.TimeNowUTC(),
	}); err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err)
	}

	return account, TokenResult{
		AccessToken:     accessToken.Token,
		RefreshToken:    refreshToken.Raw,
		AccessExpireAt:  accessToken.ExpiresAt,
		RefreshExpireAt: refreshToken.ExpiresAt,
		Type:            pkg.AuthorizationType,
	}, nil
}

func (u *usecase) VerifyEmail(ctx context.Context, emailToken string) error {
	email, err := u.deps.VerifyProvider.VerifyVerification(ctx, emailToken)
	if err != nil {
		return pkg.ErrBadRequest
	}

	if err = u.deps.UserAccount.VerifyEmail(ctx, email); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (u *usecase) ForgotPassword(ctx context.Context, email string) error {
	account, err := u.deps.UserFetch.GetByEmail(ctx, email)
	if err != nil {
		return pkg.ErrInvalidCredentials
	}
	_ = u.deps.VerifyProvider.SendPasswordReset(ctx, account.Email, account.Username)
	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, params ResetPasswordParams) error {
	email, err := u.deps.VerifyProvider.VerifyPasswordReset(ctx, params.EmailToken)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	if err = u.deps.UserAccount.ResetPassword(ctx, user.ResetPasswordParams{
		Email:       email,
		NewPassword: params.Password,
	}); err != nil {
		return pkg.OrInternalError(err, pkg.ErrInvalidCredentials)
	}

	_ = u.deps.VerifyProvider.InvalidatePasswordReset(ctx, params.EmailToken)

	return nil
}

func (u *usecase) ValidateResetPasswordToken(ctx context.Context, token string) bool {
	return u.deps.VerifyProvider.ValidateResetPasswordToken(ctx, token)
}

func (u *usecase) RefreshToken(ctx context.Context, refreshToken string) (TokenResult, error) {
	var empty TokenResult

	token, err := u.deps.TokenRepo.GetByHash(ctx, u.deps.TokenProvider.HashToken(refreshToken))
	if err != nil {
		return empty, pkg.ErrUnauthorized
	}

	if err := u.validateRefreshToken(ctx, *token); err != nil {
		return empty, err
	}

	return u.rotateToken(ctx, *token)
}

func (u *usecase) validateRefreshToken(ctx context.Context, token token.RefreshToken) error {
	isValid, err := u.deps.TokenProvider.VerifyRefreshToken(token)
	if err != nil {
		if errors.Is(err, pkg.ErrTokenRevoked) {
			// security
			_ = u.deps.TokenRepo.UpdateRevokedByUser(ctx, token.UserID)
		}
		return pkg.ErrUnauthorized
	}
	if !isValid {
		return pkg.ErrUnauthorized
	}
	return nil
}

func (u *usecase) rotateToken(ctx context.Context, oldToken token.RefreshToken) (TokenResult, error) {
	var empty TokenResult

	refreshToken := u.deps.TokenProvider.GenerateRefreshToken()
	accessToken, err := u.deps.TokenProvider.GenerateAccessToken(oldToken.UserID, oldToken.DeviceID)
	if err != nil {
		return empty, pkg.ErrInternal
	}

	if err := u.deps.TokenRepo.RotateToken(ctx, oldToken.ID, &token.RefreshToken{
		UserID:    oldToken.UserID,
		DeviceID:  oldToken.DeviceID,
		TokenHash: refreshToken.Hashed,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: pkg.TimeNowUTC(),
	}); err != nil {
		return empty, pkg.ErrInternal
	}

	return TokenResult{
		AccessToken:     accessToken.Token,
		RefreshToken:    refreshToken.Raw,
		AccessExpireAt:  accessToken.ExpiresAt,
		RefreshExpireAt: refreshToken.ExpiresAt,
		Type:            pkg.AuthorizationType,
	}, nil
}
