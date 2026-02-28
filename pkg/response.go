package pkg

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

//
// ============================================================================
// RESPONSE MODELS
// ============================================================================
//

// Response is the standard API response envelope.
// It ensures consistent response format across the entire system.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

//
// ============================================================================
// RESPONSE MODELS
// ============================================================================
//

// ErrorMsg is a custom application error wrapper.
// It allows attaching a custom message while preserving
// the original error type for errors.Is comparison.
type ErrorMsg struct {
	Err error  // sentinel error (used for errors.Is)
	Msg string // client-facing message
}

func (e ErrorMsg) Error() string {
	return e.Msg // Return client-facing message by default for simple printing
}

func (e ErrorMsg) Unwrap() error {
	return e.Err
}

func NewError(err error, msg string) error {
	return ErrorMsg{
		Err: err,
		Msg: msg,
	}
}

//
// ============================================================================
// GENERIC JSON RESPONSE HELPERS
// ============================================================================
//

func JSON(c *gin.Context, code int, msg string, data any) {
	c.JSON(code, Response{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, msg string, errs any) {
	c.AbortWithStatusJSON(code, Response{
		Code:    code,
		Message: msg,
		Errors:  errs,
	})
}

//
// ============================================================================
// GENERIC JSON RESPONSE HELPERS
// ============================================================================
//

func Success(c *gin.Context, data any) {
	JSON(c, http.StatusOK, "success", data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func SuccessWithMsg(c *gin.Context, msg string, data any) {
	JSON(c, http.StatusOK, msg, data)
}

func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, "created", data)
}

//
// ============================================================================
// STANDARD HTTP ERROR RESPONSES
// ============================================================================
//

func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg, nil)
}

func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, msg, nil)
}

func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, msg, nil)
}

func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, msg, nil)
}

func Conflict(c *gin.Context, msg string) {
	Error(c, http.StatusConflict, msg, nil)
}

func EntityTooLarge(c *gin.Context, msg string) {
	Error(c, http.StatusRequestEntityTooLarge, msg, nil)
}

func InternalError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, msg, nil)
}

//
// ============================================================================
// VALIDATION ERROR HANDLER (HTTP BINDING LAYER)
// ============================================================================
//

// HandleValidateError processes errors occurring during the HTTP request binding and validation phase.
// It categorizes errors from `StrictBindJSON` and standard validators into user-friendly JSON responses.
func HandleValidateError(c *gin.Context, err error) {
	// 1. Empty Body
	if errors.Is(err, ErrEmptyBody) {
		BadRequest(c, err.Error())
		return
	}

	// 2. JSON Format & Structure Errors (json.Decoder, e.g. malformed or has unknown fields.)
	if jsonErrors := ParseJSONError(err); len(jsonErrors) > 0 {
		Error(c, http.StatusBadRequest, "invalid request format", jsonErrors)
		return
	}

	// 3. Logical Validation Errors (binding.Validator, e.g. required, min, max, email...)
	if v := ValidateRequestError(err); v != nil {
		Error(c, http.StatusBadRequest, "invalid request payload", v.Errors)
		return
	}

	// 4. Fallback: Unknown/Generic Errors
	BadRequest(c, "invalid request payload")
}

//
// ============================================================================
// SERVICE ERROR HANDLER (DOMAIN → HTTP MAPPING LAYER)
// ============================================================================
//

// HandleServiceError maps business logic errors returned by the Service Layer
// to the appropriate HTTP status codes and JSON responses.
// It acts as a centralized error translation layer to ensure consistent API behavior.
func HandleServiceError(c *gin.Context, err error) {
	// 1. Determine Client Message and Root Cause
	var (
		clientMsg = err.Error() // Default to error message
		rootErr   = err         // Default to original error
		appErr    *ErrorMsg     // Custom wrapper check
	)

	// If wrapped with pkg.NewError, extract the custom message and the inner error
	if errors.As(err, &appErr) {
		clientMsg = appErr.Msg
		rootErr = appErr.Err
	}

	// 2. Map Root Cause to HTTP Status
	switch {
	// 401 Unauthorized
	case errors.Is(rootErr, ErrUnauthorized), errors.Is(rootErr, ErrInvalidCredentials):
		Unauthorized(c, clientMsg)

	// 403 Forbidden
	case errors.Is(rootErr, ErrForbidden):
		Forbidden(c, clientMsg)

	// 409 Conflict
	case errors.Is(rootErr, ErrAlreadyExists):
		Conflict(c, clientMsg)

	// 404 Not Found
	case errors.Is(rootErr, ErrNotFound):
		NotFound(c, clientMsg)

	// 413 Payload Too Large
	case errors.Is(rootErr, ErrFileTooLarge):
		EntityTooLarge(c, clientMsg)

	// 400 Bad Request
	case errors.Is(rootErr, ErrBadRequest),
		errors.Is(rootErr, ErrInvalidData),
		errors.Is(rootErr, ErrSamePassword):
		BadRequest(c, clientMsg)

	// 500 Internal Server Error
	default:
		Log().Errorw("internal server error", "error", err)
		InternalError(c, "an unexpected error occurred")
	}
}
