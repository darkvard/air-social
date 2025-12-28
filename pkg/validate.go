package pkg

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"

	"air-social/internal/domain"
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

func ValidateEventData(eventType string, data any) error {
	switch eventType {
	case domain.EmailVerify:
		if _, ok := data.(domain.EventEmailVerify); !ok {
			return fmt.Errorf("invalid event data: expected EventEmailVerify for type %s, got %T", eventType, data)
		}
	default:
		return fmt.Errorf("unregistered event type: %s", eventType)
	}
	return nil
}
