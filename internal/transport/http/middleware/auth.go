package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/service"
	"air-social/pkg"
)

type authContextKey string

const AuthPayloadKey authContextKey = "auth_payload"

func Auth(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get raw token string
		tokenString, err := pkg.ExtractTokenFromHeader(c)
		if err != nil {
			pkg.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Validate
		validatedToken, err := tokenService.Validate(tokenString)
		if err != nil || !validatedToken.Valid {
			pkg.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Get data
		clams, ok := validatedToken.Claims.(jwt.MapClaims)
		if !ok {
			pkg.Unauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		userID := pkg.GetInt64Claims(clams, pkg.JWTClaimSubject)
		if userID <= 0 {
			pkg.Unauthorized(c, "user id must be positive")
			c.Abort()
			return
		}

		deviceID := pkg.GetStringClaims(clams, pkg.JWTClaimDevice)

		payload := &domain.AuthPayload{
			UserID:   userID,
			DeviceID: deviceID,
		}

		// Set context
		c.Set(AuthPayloadKey, payload)
		c.Next()
	}
}

func Basic(cfg config.ServerConfig) gin.HandlerFunc {
	return gin.BasicAuth(
		gin.Accounts{
			cfg.AuthUsername: cfg.AuthPassword,
		},
	)
}

func GetAuthPayload(c *gin.Context) (*domain.AuthPayload, error) {
	value, exists := c.Get(AuthPayloadKey)
	if !exists {
		return nil, pkg.ErrUnauthorized
	}

	payload, ok := value.(*domain.AuthPayload)
	if !ok {
		return nil, pkg.ErrUnauthorized
	}

	return payload, nil
}
