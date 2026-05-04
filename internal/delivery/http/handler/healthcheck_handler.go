package handler

import (
	"log/slog"

	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthCheckHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthCheckHandler(db *gorm.DB, rdb *redis.Client) *HealthCheckHandler {
	return &HealthCheckHandler{db: db, rdb: rdb}
}

func (h *HealthCheckHandler) HealthCheck(c *echo.Context) error {
	ctx := c.Request().Context()
	sqlDB, err := h.db.DB()
	if err != nil {
		return problem.Wrap(err, problem.ErrInternalServer).WithDetail("Database connection error")
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return problem.Wrap(err, problem.ErrInternalServer).WithDetail("Database connection error")
	}

	redisStatus := "ok"
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		redisStatus = "degraded"
		slog.WarnContext(ctx, "redis health check failed", "error", err)
	}

	return c.JSON(200, map[string]any{
		"message": "Application is healthy",
		"status":  "ok",
		"checks": map[string]string{
			"database": "ok",
			"redis":    redisStatus,
		},
	})
}
