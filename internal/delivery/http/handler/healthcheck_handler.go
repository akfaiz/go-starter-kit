package handler

import (
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type HealthCheckHandler struct {
	db *bun.DB
}

func NewHealthCheckHandler(db *bun.DB) *HealthCheckHandler {
	return &HealthCheckHandler{db: db}
}

func (h *HealthCheckHandler) HealthCheck(c *echo.Context) error {
	ctx := c.Request().Context()
	if err := h.db.PingContext(ctx); err != nil {
		return problem.Wrap(err, problem.ErrInternalServer).WithDetail("Database connection error")
	}
	return c.JSON(200, map[string]string{"message": "Application is healthy", "status": "ok"})
}
