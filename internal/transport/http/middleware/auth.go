package middleware

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/service"
	"air-social/pkg"
)

type userContextKey string

const UserIDContextKey userContextKey = "userID"

func AuthMiddleware(tokenService service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get raw token string
		tokenString, err := pkg.ExtractTokenFromHeader(c)
		if err != nil {
			pkg.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Check blocked
		isBlocked, err := tokenService.IsBlocked(c.Request.Context(), tokenString)
		if err != nil {
			pkg.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		if isBlocked {
			pkg.Unauthorized(c, "token has been revoked")
			c.Abort()
			return
		}

		// Validate and get data
		validatedToken, err := tokenService.Validate(tokenString)
		if err != nil || !validatedToken.Valid {
			pkg.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		var userID int64
		if err := pkg.ExtractClaimFromToken(validatedToken, pkg.JWTClaimSubject, &userID); err != nil {
			pkg.Unauthorized(c, "invalid user identifier in token")
			c.Abort()
			return
		}
		if userID <= 0 {
			pkg.Unauthorized(c, "user id must be positive")
			c.Abort()
			return
		}

		// Set context
		c.Set(string(UserIDContextKey), userID)
		c.Next()
	}
}
