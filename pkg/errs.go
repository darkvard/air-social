package pkg

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInternal    = errors.New("internal server error")
	ErrNotFound    = errors.New("resource not found")
	ErrConflict    = errors.New("resource conflict occurred")
	ErrInvalidData = errors.New("the provided data is invalid")
	ErrDatabase    = errors.New("database error occurred")

	ErrAlreadyExists = errors.New("resource already exists")
	ErrKeyNotFound   = errors.New("key was not found")
	ErrSamePassword  = errors.New("new password must be different from current password")

	ErrSessionExpired     = errors.New("session has expired")
	ErrUnauthorized       = errors.New("authentication required")
	ErrForbidden          = errors.New("access denied")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")

	ErrFileUnsupported = errors.New("file format not supported")
	ErrFileTooLarge    = errors.New("file size exceeds limit")
	ErrFileTypeInvalid = errors.New("detected file type is invalid")
)

const (
	CodeUniqueViolation      = "23505"
	CodeForeignKeyViolation  = "23503"
	CodeCheckConstraint      = "23514"
	CodeNotNullViolation     = "23502"
	CodeSerializationFailure = "40001"
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func MapPostgresError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	switch pgCode(err) {
	case CodeUniqueViolation:
		return ErrAlreadyExists

	case CodeForeignKeyViolation:
		return ErrInvalidData

	case CodeCheckConstraint:
		return ErrInvalidData

	case CodeNotNullViolation:
		return ErrInvalidData

	case CodeSerializationFailure:
		return ErrConflict
	}

	return ErrDatabase
}

func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "html/template"):
		return true
	case strings.Contains(msg, "json:"):
		return true
	case strings.Contains(msg, "marshal"):
		return true
	case strings.Contains(msg, "unmarshal"):
		return true
	case strings.Contains(msg, "nil pointer"):
		return true
	case strings.Contains(msg, "index out of range"):
		return true
	}

	return false
}

// SkipError checks if the given error matches the target error.
// If it matches, it returns nil, otherwise, it returns the original error.
func SkipError(original error, target error) error {
	if errors.Is(original, target) {
		return nil
	}
	return original
}

// OrInternalError acts as a security filter for errors.
//
// By default, it logs the original error and returns a generic pkg.ErrInternal
// to prevent leaking sensitive system details (e.g., DB connection strings, SQL errors).
// However, if the error is present in the 'allowedErrors' list,
// it returns the original error to the client.
func OrInternalError(original error, allowed ...error) error {
	if original == nil {
		return nil
	}

	for _, err := range allowed {
		if errors.Is(original, err) {
			return original
		}
	}

	Log().Errorw("[INTERNAL ERROR]", "error", original)

	return ErrInternal
}
