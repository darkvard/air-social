package pkg

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInternal      = errors.New("internal server error")   // 500
	ErrNotFound      = errors.New("resource not found")      // 404
	ErrAlreadyExists = errors.New("resource already exists") // 409
	ErrConflict      = errors.New("conflict error")          // 409

	ErrBadRequest   = errors.New("bad request")                                          // 400
	ErrInvalidData  = errors.New("validation failed")                                    // 400
	ErrSamePassword = errors.New("new password must be different from current password") // 400

	ErrInvalidCredentials = errors.New("email or password is incorrect") // 401
	ErrUnauthorized       = errors.New("authentication required")        // 401
	ErrForbidden          = errors.New("access denied")                  // 403

	ErrFileUnsupported = errors.New("file format not supported")     // 400
	ErrFileTooLarge    = errors.New("file size exceeds limit")       // 413
	ErrFileTypeInvalid = errors.New("detected file type is invalid") // 400
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

	return err
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

	return ErrInternal
}
