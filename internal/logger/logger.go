package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	cerrors "github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultServiceName  = "go-starter-kit"
	disabledLogLevel    = slog.Level(100)
	maxErrorStackFrames = 5
)

func Init(cfg config.App, lp *sdklog.LoggerProvider) {
	threshold := resolveLogLevel(cfg.LogLevel)

	otelHandler := otelslog.NewHandler(defaultServiceName, otelslog.WithLoggerProvider(lp))
	stdoutHandler := newConsoleHandler(cfg.LogFormat, os.Stdout)

	logger := slog.New(
		slogmulti.Pipe(
			newLevelFilterMiddleware(threshold),
			newEnrichAndFormatMiddleware(),
		).Handler(slogmulti.Fanout(stdoutHandler, otelHandler)),
	)
	slog.SetDefault(logger)
}

func resolveLogLevel(raw string) slog.Level {
	switch strings.ToLower(raw) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "disabled", "off":
		return disabledLogLevel
	default:
		return slog.LevelInfo
	}
}

func newLevelFilterMiddleware(threshold slog.Level) slogmulti.Middleware {
	return slogmulti.NewEnabledInlineMiddleware(
		func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool {
			if threshold == disabledLogLevel {
				return false
			}
			if level < threshold {
				return false
			}
			return next(ctx, level)
		},
	)
}

func newEnrichAndFormatMiddleware() slogmulti.Middleware {
	return slogmulti.NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			enriched := enrichRecord(ctx, record.Clone())
			return next(ctx, formatRecordMessageAsJSON(enriched))
		},
	)
}

func enrichRecord(ctx context.Context, record slog.Record) slog.Record {
	appendTraceCorrelation(ctx, &record)
	appendErrorMetadata(&record)
	return record
}

func appendTraceCorrelation(ctx context.Context, record *slog.Record) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() || !span.SpanContext().IsValid() {
		return
	}

	spanContext := span.SpanContext()
	record.AddAttrs(
		slog.String("span_id", spanContext.SpanID().String()),
		slog.String("trace_id", spanContext.TraceID().String()),
	)
}

func appendErrorMetadata(record *slog.Record) {
	var extra []slog.Attr
	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "error" {
			return true
		}

		err := errorFromValue(attr.Value)
		if err == nil {
			return true
		}

		extra = append(extra, slog.String("error_type", telemetry.ErrorType(err)))

		stack := cerrors.GetReportableStackTrace(err)
		if stack != nil {
			extra = append(
				extra,
				slog.Any("error_stack", compactFrames(limitFrames(stack.Frames, maxErrorStackFrames))),
			)
		}
		return true
	})

	if len(extra) > 0 {
		record.AddAttrs(extra...)
	}
}

func newConsoleHandler(format string, writer io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{}
	if strings.EqualFold(format, "text") {
		return slog.NewTextHandler(writer, opts)
	}
	return slog.NewJSONHandler(writer, opts)
}

type loggerStackFrame struct {
	Function string `json:"function,omitempty"`
	Module   string `json:"module,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
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

func limitFrames(frames []sentry.Frame, maxFrames int) []sentry.Frame {
	if len(frames) == 0 {
		return nil
	}

	normalized := slices.Clone(frames)
	slices.Reverse(normalized)
	if len(normalized) > maxFrames {
		normalized = normalized[:maxFrames]
	}
	return normalized
}

func formatRecordMessageAsJSON(record slog.Record) slog.Record {
	payload := map[string]any{
		"msg": record.Message,
	}

	record.Attrs(func(attr slog.Attr) bool {
		appendJSONAttr(payload, attr)
		return true
	})

	msg, err := json.Marshal(payload)
	if err != nil {
		msg = []byte(record.Message)
	}

	inlined := slog.NewRecord(record.Time, record.Level, string(msg), record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		inlined.AddAttrs(attr)
		return true
	})

	return inlined
}

func appendJSONAttr(dest map[string]any, attr slog.Attr) {
	value := attr.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		group := map[string]any{}
		for _, child := range value.Group() {
			appendJSONAttr(group, child)
		}
		dest[attr.Key] = group
		return
	}

	dest[attr.Key] = valueToAny(value)
}

func valueToAny(value slog.Value) any {
	switch value.Kind() {
	case slog.KindGroup:
		group := map[string]any{}
		for _, child := range value.Group() {
			appendJSONAttr(group, child)
		}
		return group
	case slog.KindLogValuer:
		return valueToAny(value.Resolve())
	case slog.KindString:
		return value.String()
	case slog.KindBool:
		return value.Bool()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindUint64:
		return value.Uint64()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindDuration:
		return value.Duration()
	case slog.KindTime:
		return value.Time()
	case slog.KindAny:
		return value.Any()
	default:
		return value.String()
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
