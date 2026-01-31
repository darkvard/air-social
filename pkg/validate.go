package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

type ValidationResult struct {
	Errors []FieldError `json:"errors,omitempty"`
}

func ValidateRequestError(err error) *ValidationResult {
	var result ValidationResult

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := strings.ToLower(e.Field())
			tag := e.Tag()

			var msg string
			switch tag {
			case "required":
				msg = "is required"
			case "email":
				msg = "must be a valid email"
			case "min":
				msg = "is too short"
			case "max":
				msg = "is too long"
			default:
				msg = "is invalid"
			}

			result.Errors = append(result.Errors, FieldError{
				Field:   field,
				Message: msg,
			})
		}
		return &result
	}

	return nil
}

func ParseJSONError(err error) []FieldError {
	var jsonErr *json.UnmarshalTypeError

	if errors.As(err, &jsonErr) {
		return []FieldError{{
			Field:   jsonErr.Field,
			Message: fmt.Sprintf("expected type %s but got %s", jsonErr.Type, jsonErr.Value),
		}}
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return []FieldError{{
			Field:   "body",
			Message: "invalid json syntax",
		}}
	}

	if strings.HasPrefix(err.Error(), "json: unknown field") {
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		fieldName = strings.Trim(fieldName, "\"")

		return []FieldError{{
			Field:   fieldName,
			Message: "field is not allowed",
		}}
	}

	if errors.Is(err, io.EOF) {
		return []FieldError{{
			Field:   "body",
			Message: "request body cannot be empty",
		}}
	}

	return nil
}

// StrictBindJSON parses the Request Body into 'obj' with strict structure validation.
//
// Features:
//
//  1. Strictness: Uses decoder.DisallowUnknownFields() to reject any unknown keys.
//     This is crucial for structs using Pointers (e.g., PATCH requests) to prevent
//     silent failures when client sends typo fields (e.g., "full_nme" instead of "full_name").
//
//  2. Validation: After decoding, it executes Gin's binding validator to check rules
//     (e.g., min, max, email) on the fields that were present. It does NOT enforce
//     presence (unless 'required' tag is used), making it safe for optional/partial updates.
//
// NOTE: 'obj' MUST be a pointer to a struct.
func StrictBindJSON(c *gin.Context, obj any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	// Decode JSON
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	if IsStructNilOrEmpty(obj) {
		return ErrEmptyBody
	}

	// Validate struct (required, min, max...)
	return binding.Validator.ValidateStruct(obj)
}
