package pkg

import (
	"path/filepath"
	"strings"

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

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	return allowed[ext]
}

func ValidateImageFile(filename string) *ValidationResult {
	if !isImageFile(filename) {
		return &ValidationResult{
			Errors: []FieldError{
				{Field: "file_name", Message: "invalid file extension (allowed: .jpg, .jpeg, .png, .webp)"},
			},
		}
	}
	return nil
}
