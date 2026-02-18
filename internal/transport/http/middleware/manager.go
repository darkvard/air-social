package middleware

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	"air-social/internal/domain/auth/token"
)

type Manager struct {
	Basic         gin.HandlerFunc
	Auth          gin.HandlerFunc
	JSONOnly      gin.HandlerFunc
	MultipartOnly gin.HandlerFunc
}

func NewManager(cfg config.ServerConfig, provider token.Provider) *Manager {
	return &Manager{
		Basic:         Basic(cfg),
		Auth:          Auth(provider),
		JSONOnly:      JSONOnly(),
		MultipartOnly: MultipartOnly(),
	}
}
