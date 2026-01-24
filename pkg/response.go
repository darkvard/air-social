package pkg

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Response{
		Code:    status,
		Message: message,
		Data:    data,
	})
}

func Success(c *gin.Context, data any) {
	JSON(c, http.StatusOK, "ok", data)
}

func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, "created", data)
}

func BadRequest(c *gin.Context, msg string) {
	JSON(c, http.StatusBadRequest, msg, nil)
}

func Unauthorized(c *gin.Context, msg string) {
	JSON(c, http.StatusUnauthorized, msg, nil)
}

func Forbidden(c *gin.Context, msg string) {
	JSON(c, http.StatusForbidden, msg, nil)
}

func NotFound(c *gin.Context, msg string) {
	JSON(c, http.StatusNotFound, msg, nil)
}

func Conflict(c *gin.Context, msg string) {
	JSON(c, http.StatusConflict, msg, nil)
}

func InternalError(c *gin.Context, msg string) {
	JSON(c, http.StatusInternalServerError, msg, nil)
}

func EntityTooLarge(c *gin.Context, msg string) {
	JSON(c, http.StatusRequestEntityTooLarge, msg, nil)
}

func HandleValidateError(c *gin.Context, err error) {
	if v := ValidateRequestError(err); v != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "invalid request payload",
			"errors":  v.Errors,
		})
		return
	}
	BadRequest(c, "invalid request")
}

func HandleServiceError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidCredentials):
		Unauthorized(c, msg)

	case errors.Is(err, ErrForbidden):
		Forbidden(c, msg)

	case errors.Is(err, ErrAlreadyExists):
		Conflict(c, msg)

	case errors.Is(err, ErrNotFound):
		NotFound(c, msg)

	case errors.Is(err, ErrFileTooLarge):
		EntityTooLarge(c, msg)

	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrInvalidData), errors.Is(err, ErrSamePassword),
		errors.Is(err, ErrFileUnsupported), errors.Is(err, ErrFileTypeInvalid):
		BadRequest(c, msg)

	default:
		Log().Errorw("[SERVICE ERROR]", "error", err)
		InternalError(c, "an unexpected error occurred")
	}
}
