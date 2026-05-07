package logger

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	cerrors "github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func attrsToMap(r slog.Record) map[string]any {
	out := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		out[a.Key] = a.Value.Resolve().Any()
		return true
	})
	return out
}

func TestAppendTraceCorrelation(t *testing.T) {
	t.Run("no recording span", func(t *testing.T) {
		r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
		appendTraceCorrelation(context.Background(), &r)
		if len(attrsToMap(r)) != 0 {
			t.Fatalf("expected no attrs")
		}
	})

	t.Run("recording span", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
		ctx, span := tp.Tracer("test").Start(context.Background(), "span")
		defer span.End()

		r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
		appendTraceCorrelation(ctx, &r)
		attrs := attrsToMap(r)
		if attrs["trace_id"] == "" || attrs["span_id"] == "" {
			t.Fatalf("expected trace attrs, got %#v", attrs)
		}
	})
}

func TestAppendErrorMetadata(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelError, "msg", 0)
	err := cerrors.WithStack(errors.New("boom"))
	r.AddAttrs(slog.Any("error", err))

	appendErrorMetadata(&r)
	attrs := attrsToMap(r)
	if attrs["error_type"] == "" {
		t.Fatalf("expected error_type attr")
	}
	if _, ok := attrs["error_stack"]; !ok {
		t.Fatalf("expected error_stack attr")
	}
}

func TestErrorFromValue(t *testing.T) {
	if errorFromValue(slog.StringValue("x")) != nil {
		t.Fatalf("expected nil for non-any")
	}
	if errorFromValue(slog.AnyValue("x")) != nil {
		t.Fatalf("expected nil for non-error any")
	}
	if errorFromValue(slog.AnyValue(errors.New("x"))) == nil {
		t.Fatalf("expected error")
	}
}

func TestFrameHelpers(t *testing.T) {
	frames := []sentry.Frame{
		{Function: "", Symbol: "sym", Module: "mod", Filename: "/tmp/a.go", Lineno: 10},
		{Function: "fn", Module: "mod2", AbsPath: "/tmp/b.go", Lineno: 20},
	}

	compacted := compactFrames(frames)
	if compacted[0].Function != "sym" || compacted[1].Function != "fn" {
		t.Fatalf("unexpected compacted frames: %#v", compacted)
	}

	if limitFrames(nil, 3) != nil {
		t.Fatalf("expected nil for empty frames")
	}
	limited := limitFrames(frames, 1)
	if len(limited) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(limited))
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "a", "b"); got != "a" {
		t.Fatalf("got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestInit(t *testing.T) {
	lp := sdklog.NewLoggerProvider()
	Init(config.App{LogLevel: "debug", LogFormat: "json"}, lp)
	slog.Info("hello", "k", "v")
}
