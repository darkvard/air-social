package pkg

import (
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
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

var (
	ErrInvalidInput = errors.New("invalid_input")
	ErrInternal     = errors.New("internal_error")
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

// Convert PG error → domain error
func MapPostgresError(err error) error {
	code := pgCode(err)

	switch code {
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

	default:
		return ErrDatabase
	}
}
