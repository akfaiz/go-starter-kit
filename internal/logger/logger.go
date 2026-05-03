package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"go.opentelemetry.io/otel/trace"
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
	logger := slog.New(&customHandler{handler})
	slog.SetDefault(logger)
}

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() && span.SpanContext().IsValid() {
		spanContext := span.SpanContext()
		record.Add(slog.String("span_id", spanContext.SpanID().String()))
		record.Add(slog.String("trace_id", spanContext.TraceID().String()))
	}

	return h.Handler.Handle(ctx, record)
}
