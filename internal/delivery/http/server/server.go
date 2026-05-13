package server

import (
	"github.com/akfaiz/go-starter-kit/internal/config"
	appmiddleware "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	echoopentelemetry "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func New(cfg config.Config) *echo.Echo {
	e := echo.New()

	v := validator.New()
	e.Validator = v
	e.Binder = validator.NewBinder(e.Binder, v)
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	e.Pre(echomiddleware.RemoveTrailingSlash())
	e.Use(echoopentelemetry.NewMiddleware(cfg.Telemetry.ServiceName))
	e.Use(echomiddleware.Secure())
	e.Use(appmiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(appmiddleware.CORS(cfg.Server))
	e.Use(appmiddleware.I18n())
	e.Use(appmiddleware.RateLimiter(cfg.Server))

	return e
}
