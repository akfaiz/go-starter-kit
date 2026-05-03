package handler

import (
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type HealthCheckHandler struct {
	db *gorm.DB
}

func NewHealthCheckHandler(db *gorm.DB) *HealthCheckHandler {
	return &HealthCheckHandler{db: db}
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
	return c.JSON(200, map[string]string{"message": "Application is healthy", "status": "ok"})
}
