package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	goredis "github.com/redis/go-redis/v9"

	"air-social/pkg"
)

type HealthHandler struct {
	db     *sqlx.DB
	redis  *goredis.Client
	rabbit *amqp.Connection
}

func NewHealthHandler(db *sqlx.DB, redis *goredis.Client, rabbit *amqp.Connection) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		rabbit: rabbit,
	}
}

// HealthCheck godoc
//
//	@Summary		Health check
//	@Description	Check the health status of the application components
//	@Tags			Health
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "ok"
	dbStatus := "ok"
	redisStatus := "ok"
	rabbitStatus := "ok"

	if err := h.db.Ping(); err != nil {
		dbStatus = err.Error()
		status = "error"
	}

	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		redisStatus = err.Error()
		status = "error"
	}

	if h.rabbit.IsClosed() {
		rabbitStatus = "connection closed"
		status = "error"
	}

	pkg.Success(c, gin.H{
		"status":    status,
		"db":        dbStatus,
		"redis":     redisStatus,
		"rabbitmq":  rabbitStatus,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
