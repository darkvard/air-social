package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func JSON(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Code:    status,
		Message: message,
		Data:    data,
	})
}

func Success(c *gin.Context, data any) {
	JSON(c, http.StatusOK, "ok", data)
}

func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, "created", data)
}

func BadRequest(c *gin.Context, msg string) {
	JSON(c, http.StatusBadRequest, msg, nil)
}

func Unauthorized(c *gin.Context, msg string) {
	JSON(c, http.StatusUnauthorized, msg, nil)
}

func MissingAuthHeader(c *gin.Context) {
	Unauthorized(c, "authorization header is required")
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
	switch err {
	case ErrUnauthorized:
		Unauthorized(c, msg)
	case ErrForbidden:
		Forbidden(c, msg)
	case ErrInvalidCredentials:
		Unauthorized(c, msg)
	case ErrAlreadyExists:
		Conflict(c, msg)
	case ErrNotFound:
		NotFound(c, msg)
	case ErrInvalidData, ErrSamePassword:
		BadRequest(c, msg)
	default:
		Log().Errorw("unhandled service error", "error", err)
		InternalError(c, "an unexpected error occurred")
	}
}
