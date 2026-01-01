package pkg

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrInvalidData   = errors.New("invalid data")
	ErrDatabase      = errors.New("database error")
	ErrKeyNotFound   = errors.New("key not found")
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternal           = errors.New("internal error")
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked")
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
