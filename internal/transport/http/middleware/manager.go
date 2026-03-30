package middleware

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
	limitdomain "air-social/internal/domain/limit"
)

type Manager struct {
	Basic         gin.HandlerFunc
	Auth          gin.HandlerFunc
	JSONOnly      gin.HandlerFunc
	MultipartOnly gin.HandlerFunc

	limiter    limitdomain.RateLimiter
	limiterCfg config.RateLimiterConfig
}

func NewManager(cfg config.ServerConfig, limiterCfg config.RateLimiterConfig, provider token.Provider, limiter limitdomain.RateLimiter) Manager {
	return Manager{
		Basic:         Basic(cfg),
		Auth:          Auth(provider),
		JSONOnly:      JSONOnly(),
		MultipartOnly: MultipartOnly(),
		limiter:       limiter,
		limiterCfg:    limiterCfg,
	}
}

// ipConfig returns a looser config for IP-level protection (5x limit, same window).
// Shared IPs (office, café, dorm WiFi) need enough headroom for concurrent users.
// Example: login limit=10/account → IP limit=50 covers ~5 users logging in simultaneously.
func ipConfig(cfg config.LimiterConfig) config.LimiterConfig {
	return config.LimiterConfig{Limit: cfg.Limit * 5, Window: cfg.Window}
}

// --- Pre-auth endpoints: 2 layers (IP + email) ---
// Register only uses IP layer — per-email limit is redundant due to unique email constraint.

func (m Manager) RateLimitLoginByIP() gin.HandlerFunc {
	return RateLimitByIP(m.limiter, "login", ipConfig(m.limiterCfg.Login))
}

func (m Manager) RateLimitLoginByEmail() gin.HandlerFunc {
	return RateLimitByEmailBody(m.limiter, "login", m.limiterCfg.Login)
}

func (m Manager) RateLimitRegister() gin.HandlerFunc {
	return RateLimitByIP(m.limiter, "register", ipConfig(m.limiterCfg.Register))
}

func (m Manager) RateLimitForgotPasswordByIP() gin.HandlerFunc {
	return RateLimitByIP(m.limiter, "forgot_password", ipConfig(m.limiterCfg.ForgotPassword))
}

func (m Manager) RateLimitForgotPasswordByEmail() gin.HandlerFunc {
	return RateLimitByEmailBody(m.limiter, "forgot_password", m.limiterCfg.ForgotPassword)
}

// --- Post-auth endpoints: per-user ---

func (m Manager) RateLimitPost() gin.HandlerFunc {
	return RateLimitByUser(m.limiter, "post", m.limiterCfg.Post)
}

func (m Manager) RateLimitComment() gin.HandlerFunc {
	return RateLimitByUser(m.limiter, "comment", m.limiterCfg.Comment)
}

func (m Manager) RateLimitFollow() gin.HandlerFunc {
	return RateLimitByUser(m.limiter, "follow", m.limiterCfg.Follow)
}

func (m Manager) RateLimitMediaPresigned() gin.HandlerFunc {
	return RateLimitByUser(m.limiter, "media_presigned", m.limiterCfg.MediaPresigned)
}
