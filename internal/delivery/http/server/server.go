package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	appmiddleware "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	echoopentelemetry "github.com/labstack/echo-opentelemetry"
	echoprometheus "github.com/labstack/echo-prometheus"
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
	e.Use(echoprometheus.NewMiddleware(cfg.App.Name))
	e.Use(echomiddleware.Secure())
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

	if cfg.Server.RateLimitEnabled {
		e.Use(echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
			Skipper: echomiddleware.DefaultSkipper,
			Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
				echomiddleware.RateLimiterMemoryStoreConfig{
					Rate:      cfg.Server.RateLimitRequests,
					Burst:     cfg.Server.RateLimitBurst,
					ExpiresIn: 3 * time.Minute,
				},
			),
			IdentifierExtractor: func(c *echo.Context) (string, error) {
				return c.RealIP(), nil
			},
			ErrorHandler: func(_ *echo.Context, err error) error {
				return problem.Wrap(err, problem.ErrInternalServer)
			},
			DenyHandler: func(_ *echo.Context, _ string, _ error) error {
				return problem.ErrTooManyRequests("You have exceeded the rate limit. Please try again later.")
			},
		}))
	}

	e.GET("/metrics", echoprometheus.NewHandler())

	return e
}
