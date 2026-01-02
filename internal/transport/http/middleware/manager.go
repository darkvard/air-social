package middleware

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/config"
	"air-social/internal/service"
)

type Manager struct {
	Basic         gin.HandlerFunc
	Auth          gin.HandlerFunc
	JSONOnly      gin.HandlerFunc
	MultipartOnly gin.HandlerFunc
}

func NewManager(cfg config.ServerConfig, tokens service.TokenService) *Manager {
	return &Manager{
		Basic:         Basic(cfg),
		Auth:          Auth(tokens),
		JSONOnly:      JSONOnly(),
		MultipartOnly: MultipartOnly(),
	}
}
