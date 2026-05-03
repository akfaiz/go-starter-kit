package server

import (
	"log/slog"
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/config"
	appmiddleware "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	echoopentelemetry "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func New(cfg config.Config) *echo.Echo {
	e := echo.New()
	e.Validator = validator.New()
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Pre(echomiddleware.RemoveTrailingSlash())
	e.Use(echoopentelemetry.NewMiddleware(cfg.Telemetry.ServiceName))
	e.Use(appmiddleware.Logger(slog.Default()))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: cfg.Server.CORSOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
	}))
	e.Use(appmiddleware.I18n())
	return e
}
