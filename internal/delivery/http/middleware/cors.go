package middleware

import (
	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func CORS(cfg config.Server) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{
			"GET",
			"HEAD",
			"PUT",
			"PATCH",
			"POST",
			"DELETE",
		},
	})
}
