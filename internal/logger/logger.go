package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/config"
)

func Init(cfg config.App) {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "disabled", "off":
		level = slog.Level(100) // disable all logs
	}
	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}
	if strings.ToLower(cfg.LogFormat) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
