package auth

import (
	"context"
	"errors"

	"air-social/internal/domain/auth/token"
	"air-social/internal/domain/auth/verify"
	"air-social/internal/domain/shared"
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

	RefreshToken(ctx context.Context, refreshToken string) (TokenResult, error)
}

type Deps struct {
	TokenRepo     token.Repository
	TokenProvider token.Provider
	UserFetch     user.FetchUseCase
	UserAccount   user.AccountUseCase
	Cache         shared.Cache
}

type usecase struct {
	tokenRepo      token.Repository
	tokenProvider  token.Provider
	verifyProvider verify.Provider
	userFetch      user.FetchUseCase
	userAccount    user.AccountUseCase
	cache          shared.Cache
}

func NewUseCase(deps Deps) UseCase {
	return &usecase{
		tokenProvider: deps.TokenProvider,
		userFetch:     deps.UserFetch,
		userAccount:   deps.UserAccount,
		cache:         deps.Cache,
	}
}

func (u *usecase) Register(ctx context.Context, params RegisterParams) (*user.User, error) {
	arg := user.CreateParams{
		Email:       params.Email,
		Username:    params.Username,
		NewPassword: params.Password,
	}

	result, err := u.userAccount.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	_ = u.verifyProvider.SendVerification(ctx, arg.Email, arg.Username)

	return result, nil
}

func (u *usecase) Logout(ctx context.Context, params LogoutParams) error {
	if u.tokenProvider.IsBlacklisted(ctx, params.Token) {
		return pkg.ErrUnauthorized
	}

	var err error
	if params.IsAllDevices {
		err = u.tokenRepo.UpdateRevokedByUser(ctx, params.UserID)
	} else {
		err = u.tokenRepo.UpdateRevokedByDevice(ctx, params.UserID, params.DeviceID)
	}
	if err != nil {
		return pkg.OrInternalError(err)
	}

	u.tokenProvider.AddToBlacklist(ctx, params.Token, params.ExpiresAt)
	return nil
}

func (u *usecase) Login(ctx context.Context, params LoginParams) (*user.User, TokenResult, error) {
	var emptyUser *user.User
	var emptyToken TokenResult

	account, err := u.userAccount.Authenticate(ctx, user.AuthenticateParams{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err, pkg.ErrInvalidCredentials)
	}

	refreshToken := u.tokenProvider.GenerateRefreshToken()
	accessToken, err := u.tokenProvider.GenerateAccessToken(account.ID, params.DeviceID)
	if err != nil {
		return emptyUser, emptyToken, pkg.ErrInternal
	}

	if err = u.tokenRepo.Create(ctx, &token.RefreshToken{
		UserID:    account.ID,
		DeviceID:  params.DeviceID,
		TokenHash: refreshToken.Hashed,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: pkg.TimeNowUTC(),
	}); err != nil {
		return emptyUser, emptyToken, pkg.OrInternalError(err)
	}

	return account, TokenResult{
		AccessToken:    accessToken.Token,
		RefreshToken:   refreshToken.Raw,
		AccessExpireAt: accessToken.ExpiresAt,
		Type:           pkg.AuthorizationType,
	}, nil
}

func (u *usecase) VerifyEmail(ctx context.Context, emailToken string) error {
	email, err := u.verifyProvider.VerifyVerification(ctx, emailToken)
	if err != nil {
		return pkg.ErrBadRequest
	}

	if err = u.userAccount.VerifyEmail(ctx, email); err != nil {
		return pkg.OrInternalError(err)
	}
	return nil
}

func (u *usecase) ForgotPassword(ctx context.Context, email string) error {
	account, err := u.userFetch.GetByEmail(ctx, email)
	if err != nil {
		return pkg.ErrInvalidCredentials
	}
	_ = u.verifyProvider.SendPasswordReset(ctx, account.Email, account.Username)
	return nil
}

func (u *usecase) ResetPassword(ctx context.Context, params ResetPasswordParams) error {
	email, err := u.verifyProvider.VerifyPasswordReset(ctx, params.EmailToken)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	if err = u.userAccount.ResetPassword(ctx, user.ResetPasswordParams{
		Email:       email,
		NewPassword: params.Password,
	}); err != nil {
		return pkg.OrInternalError(err, pkg.ErrInvalidCredentials)
	}

	_ = u.verifyProvider.InvalidatePasswordReset(ctx, params.EmailToken)

	return nil
}

func (u *usecase) RefreshToken(ctx context.Context, refreshToken string) (TokenResult, error) {
	var empty TokenResult

	token, err := u.tokenRepo.GetByHash(ctx, u.tokenProvider.HashToken(refreshToken))
	if err != nil {
		return empty, pkg.ErrUnauthorized
	}

	if err := u.validateRefreshToken(ctx, *token); err != nil {
		return empty, err
	}

	return u.rotateToken(ctx, *token)
}

func (u *usecase) validateRefreshToken(ctx context.Context, token token.RefreshToken) error {
	isValid, err := u.tokenProvider.VerifyRefreshToken(token)
	if err != nil {
		if errors.Is(err, pkg.ErrTokenRevoked) {
			// security
			_ = u.tokenRepo.UpdateRevokedByUser(ctx, token.UserID)
		}
		return pkg.ErrUnauthorized
	}
	if !isValid {
		return pkg.ErrUnauthorized
	}
	return nil
}

func (u *usecase) rotateToken(ctx context.Context, token token.RefreshToken) (TokenResult, error) {
	var empty TokenResult
	if err := u.tokenRepo.UpdateRevoked(ctx, token.ID); err != nil {
		return empty, pkg.ErrInternal
	}

	refreshToken := u.tokenProvider.GenerateRefreshToken()
	accessToken, err := u.tokenProvider.GenerateAccessToken(token.UserID, token.DeviceID)
	if err != nil {
		return empty, pkg.ErrInternal
	}

	return TokenResult{
		AccessToken:    accessToken.Token,
		RefreshToken:   refreshToken.Raw,
		AccessExpireAt: accessToken.ExpiresAt,
		Type:           pkg.AuthorizationType,
	}, nil
}
