package middleware

import (
	"log/slog"
	"reflect"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogMethod:    true,
		LogURI:       true,
		LogLatency:   true,
		LogRequestID: true,
		HandleError:  true,
		Skipper: func(c *echo.Context) bool {
			skipPrefixes := []string{"/assets", "/docs", "/favicon.ico", "/metrics"}
			for _, prefix := range skipPrefixes {
				if len(c.Request().URL.Path) >= len(prefix) && c.Request().URL.Path[:len(prefix)] == prefix {
					return true
				}
			}
			return false
		},
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			status := v.Status
			if status >= 500 {
				level = slog.LevelError
			} else if status >= 400 {
				level = slog.LevelWarn
			}
			attrs := []slog.Attr{
				slog.String("request_id", v.RequestID),
				slog.Int("status", status),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Float64("latency", float64(v.Latency.Microseconds())/1000),
			}

			if v.Error != nil {
				errorAttrs := []slog.Attr{
					slog.String("msg", v.Error.Error()),
					slog.String("type", reflect.TypeOf(v.Error).String()),
				}
				stack := errors.GetReportableStackTrace(v.Error)
				if stack != nil {
					frames := stack.Frames
					slices.Reverse(frames)
					if len(frames) > 5 {
						frames = frames[:5]
					}
					errorAttrs = append(errorAttrs, slog.Any("stack", frames))
				}
				attrs = append(attrs, slog.GroupAttrs("error", errorAttrs...))
			}
			logger.LogAttrs(c.Request().Context(), level, "request", attrs...)
			return nil
		},
	})
}
