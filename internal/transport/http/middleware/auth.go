package middleware

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
	"air-social/pkg"
)

const (
	claimsKey authContextKey = "token_claims_key"
	accessKey authContextKey = "access_token_key"
)

type authContextKey string

func Auth(provider token.Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken, err := pkg.ExtractTokenFromHeader(c)
		if err != nil {
			pkg.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		claims, access, err := provider.VerifyAccessToken(jwtToken)
		if err != nil {
			pkg.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(claimsKey, claims)
		c.Set(accessKey, access)
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

func GetTokenClaims(c *gin.Context) (token.TokenClaims, error) {
	var empty token.TokenClaims
	value, exists := c.Get(claimsKey)
	if !exists {
		return empty, pkg.ErrUnauthorized
	}
	token, ok := value.(token.TokenClaims)
	if !ok {
		return empty, pkg.ErrUnauthorized
	}
	if token.UserID <= 0 {
		return empty, pkg.ErrUnauthorized
	}
	return token, nil
}

func GetAccessToken(c *gin.Context) (token.AccessTokenResult, error) {
	var empty token.AccessTokenResult
	value, exists := c.Get(accessKey)
	if !exists {
		return empty, pkg.ErrUnauthorized
	}
	result, ok := value.(token.AccessTokenResult)
	if !ok {
		return empty, pkg.ErrUnauthorized
	}
	return result, nil
}
