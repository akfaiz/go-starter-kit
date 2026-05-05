package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	cerrors "github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
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

type loggerStackFrame struct {
	Function string `json:"function,omitempty"`
	Module   string `json:"module,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

func (h *customHandler) Handle(ctx context.Context, record slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() && span.SpanContext().IsValid() {
		spanContext := span.SpanContext()
		record.Add(slog.String("span_id", spanContext.SpanID().String()))
		record.Add(slog.String("trace_id", spanContext.TraceID().String()))
	}
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "error" {
			return true
		}
		err := errorFromValue(attr.Value)
		if err == nil {
			return true
		}
		record.Add(slog.String("error_type", telemetry.ErrorType(err)))

		stack := cerrors.GetReportableStackTrace(err)
		if stack == nil {
			return true
		}

		frames := slices.Clone(stack.Frames)
		slices.Reverse(frames)
		if len(frames) > 5 {
			frames = frames[:5]
		}
		record.Add(slog.Any("error_stack", compactFrames(frames)))
		return false
	})

	return h.Handler.Handle(ctx, record)
}

func errorFromValue(value slog.Value) error {
	resolved := value.Resolve()
	if resolved.Kind() != slog.KindAny {
		return nil
	}

	err, ok := resolved.Any().(error)
	if !ok {
		return nil
	}

	return err
}

func compactFrames(frames []sentry.Frame) []loggerStackFrame {
	out := make([]loggerStackFrame, 0, len(frames))
	for _, frame := range frames {
		out = append(out, loggerStackFrame{
			Function: firstNonEmpty(frame.Function, frame.Symbol),
			Module:   frame.Module,
			File:     filepath.Base(firstNonEmpty(frame.Filename, frame.AbsPath)),
			Line:     frame.Lineno,
		})
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
