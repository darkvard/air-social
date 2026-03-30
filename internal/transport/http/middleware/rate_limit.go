package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	limitdomain "air-social/internal/domain/limit"
	"air-social/pkg"
)

// RateLimitByIP limits requests based on client IP.
// Used as Layer 1 for pre-auth endpoints (login, register, forgot-password).
func RateLimitByIP(limiter limitdomain.RateLimiter, rule string, cfg config.LimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		applyRateLimit(c, limiter, buildKey(rule, "ip", c.ClientIP()), cfg)
	}
}

// RateLimitByEmailBody limits requests based on the "email" field in the JSON body.
// Used as Layer 2 for pre-auth endpoints — protects a specific account regardless of IP.
// Falls through silently if email cannot be extracted (body parse error, missing field).
func RateLimitByEmailBody(limiter limitdomain.RateLimiter, rule string, cfg config.LimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		email, err := peekJSONField(c, "email")
		if err != nil || email == "" {
			return
		}
		applyRateLimit(c, limiter, buildKey(rule, "email", email), cfg)
	}
}

// RateLimitByUser limits requests based on authenticated userID.
// Must be placed after the Auth middleware in the handler chain.
func RateLimitByUser(limiter limitdomain.RateLimiter, rule string, cfg config.LimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := GetTokenClaims(c)
		if err != nil {
			return
		}
		applyRateLimit(c, limiter, buildKey(rule, "user", claims.UserID), cfg)
	}
}

// applyRateLimit is the shared core logic for all rate limit middlewares.
// Fail-open: if Redis errors, the request is allowed through (logged but not blocked).
func applyRateLimit(c *gin.Context, limiter limitdomain.RateLimiter, key string, cfg config.LimiterConfig) {
	result, err := limiter.Allow(c.Request.Context(), key, cfg.Limit, cfg.Window)
	if err != nil {
		return
	}

	c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

	if !result.Allowed {
		retryAfter := max(int64(time.Until(result.ResetAt).Seconds()), 0)
		c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
		pkg.Error(c, http.StatusTooManyRequests, "too many requests", nil)
		return
	}
}

// peekJSONField reads a single field from the JSON body without consuming it.
// Restores the body so subsequent handlers can read it normally.
func peekJSONField(c *gin.Context, field string) (string, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	val, _ := data[field].(string)
	return val, nil
}

// buildKey constructs a namespaced Redis key for rate limiting.
// Format: "social:rl:{rule}:{identifierType}:{identifierValue}"
// Example: buildKey("login", "ip", "1.2.3.4") → "social:rl:login:ip:1.2.3.4"
func buildKey(rule, identifierType string, identifierValue any) string {
	return fmt.Sprintf("social:rl:%s:%s:%v", rule, identifierType, identifierValue)
}
