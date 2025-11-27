package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "OK",
		Data:    data,
	})
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{
		Code:    http.StatusBadRequest,
		Message: msg,
	})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    http.StatusUnauthorized,
		Message: msg,
	})
}

func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    http.StatusForbidden,
		Message: msg,
	})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code:    http.StatusNotFound,
		Message: msg,
	})
}

func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: msg,
	})
}

func Respond(c *gin.Context, data Response) {
	c.JSON(http.StatusOK, data)
}
