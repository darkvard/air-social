package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"air-social/pkg"
)

const (
	headerContentType    = "Content-Type"
	contentTypeJSON      = "application/json"
	contentTypeMultipart = "multipart/form-data"
)

func isMethodChange(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

func JSONOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isMethodChange(c.Request.Method) {
			ct := c.GetHeader(headerContentType)
			if !strings.HasPrefix(ct, contentTypeJSON) {
				pkg.JSON(c, http.StatusUnsupportedMediaType, "Content-Type must be '"+contentTypeJSON+"'", nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func MultipartOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isMethodChange(c.Request.Method) {
			ct := c.GetHeader(headerContentType)
			if !strings.HasPrefix(ct, contentTypeMultipart) {
				pkg.JSON(c, http.StatusUnsupportedMediaType, "Content-Type must be '"+contentTypeMultipart+"'", nil)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
