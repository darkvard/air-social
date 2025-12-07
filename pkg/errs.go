package pkg

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("not_found")
	ErrAlreadyExists = errors.New("already_exists")
	ErrConflict      = errors.New("conflict")
	ErrInvalidData   = errors.New("invalid_data")
	ErrDatabase      = errors.New("database_error")
)

var (
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid_input")
	ErrInternal           = errors.New("internal_error")
)

var (
	ErrKeyNotFound = errors.New("key_not_found")
	ErrJWTInvalid  = errors.New("invalid_jwt")
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
