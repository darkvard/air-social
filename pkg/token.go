package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
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
	JWTClaimID        = "jti"
	JWTClaimAudience  = "aud"
	JWTClaimIssuer    = "iss"
	JWTClaimIssuedAt  = "iat"
	JWTClaimNotBefore = "nbf"
	JWTClaimExpiresAt = "exp"
)

// ExtractTokenFromHeader extracts the token from the Authorization header.
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

func ExtractClaimFromString(tokenString string, claimKey string, dest any) error {
    parser := jwt.NewParser(jwt.WithoutClaimsValidation())
    claims := jwt.MapClaims{}

    if _, _, err := parser.ParseUnverified(tokenString, &claims); err != nil {
        return err
    }

    claimValue, ok := claims[claimKey]
    if !ok {
        return errors.New("claim not found")
    }

    data, err := json.Marshal(claimValue)
    if err != nil {
        return fmt.Errorf("failed to marshal claim: %w", err)
    }
    if err := json.Unmarshal(data, dest); err != nil {
        return fmt.Errorf("failed to unmarshal claim into destination type: %w", err)
    }

    return nil
}

func ExtractClaimFromToken(token *jwt.Token, claimKey string, dest any) error {
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return errors.New("invalid token claims type")
    }

    claimValue, ok := claims[claimKey]
    if !ok {
        return fmt.Errorf("claim '%s' not found", claimKey)
    }

    data, err := json.Marshal(claimValue)
    if err != nil {
        return fmt.Errorf("failed to marshal claim: %w", err)
    }

    if err := json.Unmarshal(data, dest); err != nil {
        return fmt.Errorf("failed to unmarshal claim into destination: %w", err)
    }

    return nil
}