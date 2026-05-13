package middleware

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func RateLimiter(cfg config.Server) echo.MiddlewareFunc {
	if !cfg.RateLimitEnabled {
		return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				return next(c)
			}
		})
	}

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      cfg.RateLimitRequests,
				Burst:     cfg.RateLimitBurst,
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
	})
}
