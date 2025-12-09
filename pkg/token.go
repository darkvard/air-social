package pkg

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthorizationHeaderKey = "Authorization"
	AuthorizationType      = "Bearer"
)

// Standard JWT claims
const (
	JWTClaimSubject   = "sub"
	JWTClaimDevice    = "dev"
	JWTClaimAudience  = "aud"
	JWTClaimIssuer    = "iss"
	JWTClaimIssuedAt  = "iat"
	JWTClaimNotBefore = "nbf"
	JWTClaimExpiresAt = "exp"
)

func ExtractTokenFromHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader(AuthorizationHeaderKey)
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != AuthorizationType {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

func GetStringClaims(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func GetInt64Claims(claim jwt.MapClaims, key string) int64 {
	if val, ok := claim[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return 0
}
