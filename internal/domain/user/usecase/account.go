package usecase

import (
	"context"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"

	"air-social/internal/domain/user"
	"air-social/pkg"
)

type AccountDeps struct {
	Repo  user.Repository
	Cache user.Cache
}

type accountUseCase struct {
	deps AccountDeps
}

func NewAccountUseCase(d AccountDeps) *accountUseCase {
	return &accountUseCase{deps: d}
}

func (u *accountUseCase) VerifyEmail(ctx context.Context, email string) error {
	user, err := u.deps.Repo.GetByEmail(ctx, email)
	if err != nil {
		return pkg.OrInternalError(err, pkg.ErrNotFound)
	}
	if err := u.deps.Repo.UpdateVerified(ctx, user.ID, true); err != nil {
		return pkg.ErrInternal
	}
	_ = u.deps.Cache.Invalidate(ctx, user.ID)
	return nil
}

func (u *accountUseCase) CreateUser(ctx context.Context, params user.CreateParams) (*user.User, error) {
	hashed, err := hashPassword(params.NewPassword)
	if err != nil {
		return nil, pkg.ErrInternal
	}
	user := &user.User{
		Email:        params.Email,
		Username:     params.Username,
		PasswordHash: hashed,
	}
	if err := u.deps.Repo.Create(ctx, user); err != nil {
		return nil, pkg.OrInternalError(err, pkg.ErrAlreadyExists)
	}
	return user, nil
}

func (u *accountUseCase) ChangePassword(ctx context.Context, params user.ChangePasswordParams) error {
	user, err := u.deps.Repo.GetByID(ctx, params.UserID)
	if err != nil {
		return pkg.ErrInvalidCredentials
	}

	if params.NewPassword == params.CurrentPassword {
		return pkg.ErrSamePassword
	}
	if !verifyPassword(params.CurrentPassword, user.PasswordHash) {
		return pkg.ErrInvalidCredentials
	}

	hashed, err := hashPassword(params.NewPassword)
	if err != nil {
		return pkg.ErrInternal
	}
	return pkg.OrInternalError(
		u.deps.Repo.UpdatePassword(ctx, user.ID, hashed),
	)
}

func (u *accountUseCase) ResetPassword(ctx context.Context, params user.ResetPasswordParams) error {
	user, err := u.deps.Repo.GetByEmail(ctx, params.Email)
	if err != nil {
		return pkg.ErrInvalidCredentials
	}

	hashed, err := hashPassword(params.NewPassword)
	if err != nil {
		return pkg.ErrInternal
	}
	return pkg.OrInternalError(
		u.deps.Repo.UpdatePassword(ctx, user.ID, hashed),
	)
}

func (u *accountUseCase) Authenticate(ctx context.Context, params user.AuthenticateParams) (*user.User, error) {
	user, err := u.deps.Repo.GetByEmail(ctx, params.Email)
	if err != nil {
		return nil, pkg.ErrInvalidCredentials
	}

	if !verifyPassword(params.Password, user.PasswordHash) {
		return nil, pkg.ErrInvalidCredentials
	}

	return user, nil
}

// hashPassword generates a bcrypt hash of the password using the default cost.
//
// To circumvent bcrypt's 72-byte input truncation limit, the password is
// pre-hashed using SHA-256 before being passed to bcrypt. This ensures
// passwords of any length are securely handled.
func hashPassword(plainText string) (string, error) {
	// SHA-256 produces a fixed 32-byte hash, safe for bcrypt.
	sha := sha256.Sum256([]byte(plainText))
	hash, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(plainPassword, hashPassword string) bool {
	sha := sha256.Sum256([]byte(plainPassword))
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), sha[:])
	return err == nil
}
